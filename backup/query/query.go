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

package query

import "github.com/sjy-dv/nnv/backup/document"

// Query represents a generic query which is submitted to a specific collection.
type Query struct {
	collection string
	criteria   Criteria
	limit      int
	skip       int
	sortOpts   []SortOption
}

// NewQuery simply returns the collection with the supplied name. Use it to initialize a new query.
func NewQuery(collection string) *Query {
	return &Query{
		collection: collection,
		criteria:   nil,
		limit:      -1,
		skip:       0,
		sortOpts:   nil,
	}
}

func (q *Query) copy() *Query {
	return &Query{
		collection: q.collection,
		criteria:   q.criteria,
		limit:      q.limit,
		skip:       q.skip,
		sortOpts:   q.sortOpts,
	}
}

func (q *Query) satisfy(doc *document.Document) bool {
	if q.criteria == nil {
		return true
	}
	return q.criteria.Satisfy(doc)
}

// MatchFunc selects all the documents which satisfy the supplied predicate function.
func (q *Query) MatchFunc(p func(doc *document.Document) bool) *Query {
	return q.Where(newCriteria(FunctionOp, "", p))
}

// Where returns a new Query which select all the documents fulfilling the provided Criteria.
func (q *Query) Where(c Criteria) *Query {
	newQuery := q.copy()
	newQuery.criteria = c
	return newQuery
}

// Skip sets the query so that the first n documents of the result set are discardedocument.
func (q *Query) Skip(n int) *Query {
	if n >= 0 {
		newQuery := q.copy()
		newQuery.skip = n
		return newQuery
	}
	return q
}

func (q *Query) Limit(n int) *Query {
	newQuery := q.copy()
	newQuery.limit = n
	return newQuery
}

type SortOption struct {
	Field     string
	Direction int
}

func normalizeSortOptions(opts []SortOption) []SortOption {
	normOpts := make([]SortOption, 0, len(opts))
	for _, opt := range opts {
		if opt.Direction >= 0 {
			normOpts = append(normOpts, SortOption{Field: opt.Field, Direction: 1})
		} else {
			normOpts = append(normOpts, SortOption{Field: opt.Field, Direction: -1})
		}
	}
	return normOpts
}

// Sort sets the query so that the returned documents are sorted according list of options.
func (q *Query) Sort(opts ...SortOption) *Query {
	if len(opts) == 0 { // by default, documents are sorted documents by "_id" field
		opts = []SortOption{{Field: document.ObjectIdField, Direction: 1}}
	} else {
		opts = normalizeSortOptions(opts)
	}

	newQuery := q.copy()
	newQuery.sortOpts = opts
	return newQuery
}

func (q *Query) Collection() string {
	return q.collection
}

func (q *Query) Criteria() Criteria {
	return q.criteria
}

func (q *Query) GetLimit() int {
	return q.limit
}

func (q *Query) GetSkip() int {
	return q.skip
}

func (q *Query) SortOptions() []SortOption {
	return q.sortOpts
}
