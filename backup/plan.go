// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package backup

import (
	"sort"

	"github.com/sjy-dv/nnv/backup/document"
	"github.com/sjy-dv/nnv/backup/index"
	"github.com/sjy-dv/nnv/backup/internal"
	"github.com/sjy-dv/nnv/backup/query"
	"github.com/sjy-dv/nnv/backup/store"
)

type planNode interface {
	SetNext(next planNode)
	NextNode() planNode
	Callback(doc *document.Document) error
	Finish() error
}

type inputNode interface {
	planNode
	Run(tx store.Tx) error
}

type planNodeBase struct {
	next planNode
}

func (nd *planNodeBase) NextNode() planNode {
	return nd.next
}

func (nd *planNodeBase) SetNext(next planNode) {
	nd.next = next
}

func (nd *planNodeBase) CallNext(doc *document.Document) error {
	if nd.next != nil {
		return nd.next.Callback(doc)
	}
	return nil
}

func (nd *planNodeBase) Callback(doc *document.Document) error {
	return nil
}

func (nd *planNodeBase) Finish() error {
	return nil
}

type iterNode struct {
	planNodeBase
	filter     query.Criteria
	collection string

	//vRange     *valueRange
	//index      RangeIndex

	idxQuery index.Query
	//iterIndexReverse bool
}

func (nd *iterNode) iterateFullCollection(tx store.Tx) error {
	prefix := []byte(getDocumentKeyPrefix(nd.collection))
	return iteratePrefix(prefix, tx, func(item store.Item) error {
		doc, err := document.Decode(item.Value)
		if err != nil {
			return err
		}

		if nd.filter == nil || nd.filter.Satisfy(doc) {
			return nd.CallNext(doc)
		}

		return nil
	})
}

func (nd *iterNode) iterateIndex(tx store.Tx) error {
	iterFunc := func(docId string) error {
		doc, err := getDocumentById(nd.collection, docId, tx)

		if err != nil || doc == nil {
			// doc == nil when index record expires after document record
			return err
		}

		if nd.filter == nil || nd.filter.Satisfy(doc) {
			return nd.CallNext(doc)
		}
		return nil
	}

	err := nd.idxQuery.Run(iterFunc)
	return err
}

func (nd *iterNode) Run(tx store.Tx) error {
	if nd.idxQuery != nil {
		return nd.iterateIndex(tx)
	}
	return nd.iterateFullCollection(tx)
}

func getIndexQueries(q *query.Query, indexes []index.Index) []index.Query {
	if q.Criteria() == nil || len(indexes) == 0 {
		return nil
	}

	info := make(map[string]*index.Info)
	for _, idx := range indexes {
		info[idx.Field()] = &index.Info{
			Field: idx.Field(),
			Type:  idx.Type(),
		}
	}

	c := q.Criteria().Accept(&NotFlattenVisitor{}).(query.Criteria)
	selectedFields := c.Accept(&IndexSelectVisitor{
		Fields: info,
	}).([]*index.Info)

	if len(selectedFields) == 0 {
		return nil
	}

	indexesMap := make(map[string]index.Index)
	for _, idx := range indexes {
		indexesMap[idx.Field()] = idx
	}

	fieldRanges := c.Accept(NewFieldRangeVisitor([]string{selectedFields[0].Field})).(map[string]*index.Range)

	queries := make([]index.Query, 0)
	for field, vRange := range fieldRanges {
		queries = append(queries, &index.RangeIndexQuery{
			Range: vRange,
			Idx:   indexesMap[field].(index.RangeIndex),
		})
	}
	return queries
}

func tryToSelectIndex(q *query.Query, indexes []index.Index) (*iterNode, bool) {
	indexQueries := getIndexQueries(q, indexes)
	if len(indexQueries) == 1 {
		outputSorted := false

		idxQuery := indexQueries[0]

		if rangeQuery, ok := idxQuery.(*index.RangeIndexQuery); ok {
			if len(q.SortOptions()) == 1 && q.SortOptions()[0].Field == rangeQuery.Idx.Field() {
				rangeQuery.Reverse = q.SortOptions()[0].Direction < 0
				outputSorted = true
			}
		}

		return &iterNode{
			idxQuery:   idxQuery,
			filter:     q.Criteria(),
			collection: q.Collection(),
		}, outputSorted
	}

	if len(q.SortOptions()) == 1 {
		for _, idx := range indexes {
			if idx.Type() == index.SingleField && idx.Field() == q.SortOptions()[0].Field {
				return &iterNode{
					filter:     q.Criteria(),
					collection: q.Collection(),
					idxQuery: &index.RangeIndexQuery{
						Range:   nil,
						Idx:     idx.(index.RangeIndex),
						Reverse: q.SortOptions()[0].Direction < 0,
					},
				}, true
			}
		}
	}
	return nil, false
}

type skipLimitNode struct {
	planNodeBase
	skipped  int
	consumed int
	skip     int
	limit    int
}

func (nd *skipLimitNode) Callback(doc *document.Document) error {
	if nd.skipped < nd.skip {
		nd.skipped++
		return nil
	}

	if nd.limit < 0 || (nd.limit >= 0 && nd.consumed < nd.limit) {
		nd.consumed++
		return nd.CallNext(doc)
	}
	return internal.ErrStopIteration
}

type sortNode struct {
	planNodeBase
	opts []query.SortOption
	docs []*document.Document
}

func (nd *sortNode) Callback(doc *document.Document) error {
	if nd.docs == nil {
		nd.docs = make([]*document.Document, 0)
	}
	nd.docs = append(nd.docs, doc)
	return nil
}

func (nd *sortNode) Finish() error {
	if nd.docs != nil {
		sort.Slice(nd.docs, func(i, j int) bool {
			return compareDocuments(nd.docs[i], nd.docs[j], nd.opts) < 0
		})

		for _, doc := range nd.docs {
			nd.CallNext(doc)
		}
	}
	return nil
}

func buildQueryPlan(q *query.Query, indexes []index.Index, outputNode planNode) inputNode {
	var inputNode inputNode
	var prevNode planNode

	itNode, isOutputSorted := tryToSelectIndex(q, indexes)
	if itNode == nil {
		itNode = &iterNode{
			filter:     q.Criteria(),
			collection: q.Collection(),
		}
	}
	inputNode = itNode
	prevNode = itNode

	//isOutputSorted := (len(q.sortOpts) == 1 && itNode.index != nil && itNode.index.Field() == q.sortOpts[0].Field)
	if len(q.SortOptions()) > 0 && !isOutputSorted {
		nd := &sortNode{opts: q.SortOptions()}
		prevNode.SetNext(nd)
		prevNode = nd
	}

	//log.Println("output sorted: ", len(q.SortOptions()) > 0 && !isOutputSorted)

	if q.GetSkip() > 0 || q.GetLimit() >= 0 {
		nd := &skipLimitNode{skipped: 0, consumed: 0, skip: q.GetSkip(), limit: q.GetLimit()}
		prevNode.SetNext(nd)
		prevNode = nd
	}

	prevNode.SetNext(outputNode)

	return inputNode
}

func execPlan(nd inputNode, tx store.Tx) error {
	if err := nd.Run(tx); err != nil {
		return err
	}

	for curr := nd.(planNode); curr != nil; curr = curr.NextNode() {
		if err := curr.Finish(); err != nil {
			return err
		}
	}
	return nil
}

type consumerNode struct {
	planNodeBase
	consumer docConsumer
}

func (nd *consumerNode) Callback(doc *document.Document) error {
	return nd.consumer(doc)
}

func compareDocuments(first *document.Document, second *document.Document, sortOpts []query.SortOption) int {
	for _, opt := range sortOpts {
		field := opt.Field
		direction := opt.Direction

		firstHas := first.Has(field)
		secondHas := second.Has(field)

		if !firstHas && secondHas {
			return -direction
		}

		if firstHas && !secondHas {
			return direction
		}

		if firstHas && secondHas {
			res := internal.Compare(first.Get(field), second.Get(field))
			if res != 0 {
				return res * direction
			}
		}
	}
	return 0
}
