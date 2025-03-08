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

package main

import (
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/blevesearch/bleve/v2/analysis"
	_ "github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/registry"
)

// Token은 문서에서 추출된 토큰을 나타냅니다.
type Token struct {
	Term string
}

// Analyser 인터페이스는 텍스트 분석을 추상화합니다.
type Analyser interface {
	Analyse(text string) ([]Token, error)
}

// bleveAnalyser는 Bleve의 분석기를 래핑합니다.
type bleveAnalyser struct {
	analyzer analysis.Analyzer
}

// newBleveAnalyser는 지정한 이름의 분석기를 생성합니다.
func newBleveAnalyser(name string) (*bleveAnalyser, error) {
	cache := registry.NewCache()
	a, err := cache.AnalyzerNamed(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyser %s: %w", name, err)
	}
	return &bleveAnalyser{analyzer: a}, nil
}

// Analyse는 입력 텍스트를 토큰화합니다.
func (b *bleveAnalyser) Analyse(text string) ([]Token, error) {
	bleveTokens := b.analyzer.Analyze([]byte(text))
	tokens := make([]Token, len(bleveTokens))
	for i, t := range bleveTokens {
		tokens[i] = Token{Term: string(t.Term)}
	}
	return tokens, nil
}

// Document는 인덱싱할 문서를 나타냅니다.
type Document struct {
	ID      uint64
	Title   string
	Content string
}

// docInfo는 각 문서의 분석 결과를 저장합니다.
type docInfo struct {
	TitleFreq     map[string]int // Title의 토큰 빈도
	TitleLength   int            // Title의 토큰 총 개수
	ContentFreq   map[string]int // Content의 토큰 빈도
	ContentLength int            // Content의 토큰 총 개수
	Title         string         // 원본 Title
	Content       string         // 원본 Content
}

// SearchResult는 검색 결과(문서 ID, 점수, 원본 필드)를 나타냅니다.
type SearchResult struct {
	DocID   uint64
	Score   float64
	Title   string
	Content string
}

// Index는 Title과 Content 필드를 별도로 인덱싱하는 메모리 내 텍스트 인덱스입니다.
type Index struct {
	analyser        Analyser
	docs            map[uint64]docInfo
	invertedTitle   map[string]map[uint64]int // term -> docID -> 빈도수 (Title)
	invertedContent map[string]map[uint64]int // term -> docID -> 빈도수 (Content)
	docCount        uint64
}

// NewIndex는 지정한 분석기를 사용하여 새 인덱스를 생성합니다.
func NewIndex(analyserName string) (*Index, error) {
	a, err := newBleveAnalyser(analyserName)
	if err != nil {
		return nil, err
	}
	return &Index{
		analyser:        a,
		docs:            make(map[uint64]docInfo),
		invertedTitle:   make(map[string]map[uint64]int),
		invertedContent: make(map[string]map[uint64]int),
	}, nil
}

// Insert는 Document를 분석하여 인덱스에 추가합니다.
func (idx *Index) Insert(doc Document) error {
	// Title 분석
	titleTokens, err := idx.analyser.Analyse(doc.Title)
	if err != nil {
		return err
	}
	titleFreq := make(map[string]int)
	for _, token := range titleTokens {
		titleFreq[token.Term]++
	}

	// Content 분석
	contentTokens, err := idx.analyser.Analyse(doc.Content)
	if err != nil {
		return err
	}
	contentFreq := make(map[string]int)
	for _, token := range contentTokens {
		contentFreq[token.Term]++
	}

	info := docInfo{
		TitleFreq:     titleFreq,
		TitleLength:   len(titleTokens),
		ContentFreq:   contentFreq,
		ContentLength: len(contentTokens),
		Title:         doc.Title,
		Content:       doc.Content,
	}
	idx.docs[doc.ID] = info
	idx.docCount++

	// 역색인 업데이트 (Title)
	for term, count := range titleFreq {
		if idx.invertedTitle[term] == nil {
			idx.invertedTitle[term] = make(map[uint64]int)
		}
		idx.invertedTitle[term][doc.ID] = count
	}
	// 역색인 업데이트 (Content)
	for term, count := range contentFreq {
		if idx.invertedContent[term] == nil {
			idx.invertedContent[term] = make(map[uint64]int)
		}
		idx.invertedContent[term][doc.ID] = count
	}
	return nil
}

// Search는 주어진 query를 field("title", "content", "both")에 따라 검색합니다.
func (idx *Index) Search(query string, field string) ([]SearchResult, error) {
	// 쿼리 분석
	tokens, err := idx.analyser.Analyse(query)
	if err != nil {
		return nil, err
	}
	// 중복 제거된 쿼리 단어 집합
	queryTerms := make(map[string]struct{})
	for _, token := range tokens {
		queryTerms[token.Term] = struct{}{}
	}

	// 후보 문서 수집
	candidateDocs := make(map[uint64]struct{})
	switch field {
	case "title":
		for term := range queryTerms {
			if docs, ok := idx.invertedTitle[term]; ok {
				for docID := range docs {
					candidateDocs[docID] = struct{}{}
				}
			}
		}
	case "content":
		for term := range queryTerms {
			if docs, ok := idx.invertedContent[term]; ok {
				for docID := range docs {
					candidateDocs[docID] = struct{}{}
				}
			}
		}
	case "both":
		for term := range queryTerms {
			if docs, ok := idx.invertedTitle[term]; ok {
				for docID := range docs {
					candidateDocs[docID] = struct{}{}
				}
			}
			if docs, ok := idx.invertedContent[term]; ok {
				for docID := range docs {
					candidateDocs[docID] = struct{}{}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}

	var results []SearchResult
	// 각 후보 문서에 대해 TF‑IDF 점수 계산
	for docID := range candidateDocs {
		info := idx.docs[docID]
		score := 0.0
		// Title 점수 계산
		if field == "title" || field == "both" {
			for term := range queryTerms {
				freq := info.TitleFreq[term]
				tf := 0.0
				if info.TitleLength > 0 {
					tf = float64(freq) / float64(info.TitleLength)
				}
				docFreq := 0
				if docs, ok := idx.invertedTitle[term]; ok {
					docFreq = len(docs)
				}
				idf := math.Log10(float64(idx.docCount) / float64(docFreq+1))
				score += tf * idf
			}
		}
		// Content 점수 계산
		if field == "content" || field == "both" {
			contentScore := 0.0
			for term := range queryTerms {
				freq := info.ContentFreq[term]
				tf := 0.0
				if info.ContentLength > 0 {
					tf = float64(freq) / float64(info.ContentLength)
				}
				docFreq := 0
				if docs, ok := idx.invertedContent[term]; ok {
					docFreq = len(docs)
				}
				idf := math.Log10(float64(idx.docCount) / float64(docFreq+1))
				contentScore += tf * idf
			}
			score += contentScore
		}
		results = append(results, SearchResult{
			DocID:   docID,
			Score:   score, // 아직 raw score
			Title:   info.Title,
			Content: info.Content,
		})
	}

	// 점수 내림차순 정렬
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// **정규화: 최대 점수를 찾아서 0~100 점 범위로 변환**
	var maxScore float64 = 0
	for _, res := range results {
		if res.Score > maxScore {
			maxScore = res.Score
		}
	}
	if maxScore > 0 {
		for i, res := range results {
			results[i].Score = (res.Score / maxScore) * 100
		}
	}

	return results, nil
}

func main() {
	// "standard" 분석기를 사용하여 인덱스 생성
	index, err := NewIndex("standard")
	if err != nil {
		log.Fatalf("인덱스 생성 실패: %v", err)
	}

	// Title과 Content를 가진 한국어 샘플 문서 5개
	docs := []Document{
		{ID: 1, Title: "안녕하세요", Content: "이것은 한국어 텍스트 검색 예제입니다."},
		{ID: 2, Title: "고양이와 강아지", Content: "고양이와 강아지는 모두 귀엽습니다."},
		{ID: 3, Title: "Bleve와 Golang", Content: "Bleve를 사용하여 Golang으로 인덱싱과 검색을 구현합니다."},
		{ID: 4, Title: "날씨 정보", Content: "오늘의 날씨는 맑고 화창합니다."},
		{ID: 5, Title: "프로그래밍", Content: "프로그래밍은 창의적인 문제 해결을 요구합니다."},
	}

	// 문서 인덱싱
	for _, doc := range docs {
		if err := index.Insert(doc); err != nil {
			log.Fatalf("문서 인덱싱 실패 (ID: %d): %v", doc.ID, err)
		}
	}

	// 검색할 쿼리와 필드 지정 예시
	testCases := []struct {
		query string
		field string
	}{
		{"검색", "both"},
		{"Bleve", "title"},
		{"한국어", "both"},
		{"날씨", "content"},
		{"강아지", "both"},
	}

	for _, tc := range testCases {
		fmt.Printf("검색어: %q, 필드: %q\n", tc.query, tc.field)
		results, err := index.Search(tc.query, tc.field)
		if err != nil {
			log.Fatalf("검색 실패: %v", err)
		}
		if len(results) == 0 {
			fmt.Println("검색 결과가 없습니다.")
		} else {
			for _, res := range results {
				fmt.Printf("DocID: %d, Score: %.4f\n  Title: %s\n  Content: %s\n",
					res.DocID, res.Score, res.Title, res.Content)
			}
		}
		fmt.Println("----------")
	}
}
