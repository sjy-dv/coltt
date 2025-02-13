package playground_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/blevesearch/bleve/v2"
)

// 인덱싱할 문서 구조체 정의
type Document struct {
	Title   string
	Content string
}

func TestBleveIndex(t *testing.T) {
	// 1. 인덱스 생성
	indexPath := "example.bleve"

	// 새로운 인덱스 매핑 생성 (필요에 따라 커스터마이징 가능)
	indexMapping := bleve.NewIndexMapping()

	// 인덱스 생성 (이미 같은 경로에 인덱스가 있다면 에러가 발생할 수 있으므로 주의)
	index, err := bleve.New(indexPath, indexMapping)
	if err != nil {
		log.Fatal("인덱스 생성 에러:", err)
	}
	fmt.Println("인덱스 생성 완료")

	// 2. 데이터 추가 (문서 인덱싱)
	doc1 := Document{
		Title:   "첫 번째 문서",
		Content: "이것은 첫 번째 테스트 문서입니다.",
	}
	err = index.Index("doc1", doc1)
	if err != nil {
		log.Fatal("문서 인덱싱 에러:", err)
	}
	fmt.Println("문서 추가 완료 (doc1)")

	// 3. 데이터 업데이트 (동일한 ID를 사용하여 업데이트)
	// 기존 "doc1" 문서를 새로운 내용으로 업데이트합니다.
	doc1Updated := Document{
		Title:   "xox 업데이트트",
		Content: "이 문서는 업데이트된 내용을 포함합니다.",
	}
	err = index.Index("doc1", doc1Updated)
	if err != nil {
		log.Fatal("문서 업데이트 에러:", err)
	}
	fmt.Println("문서 업데이트 완료 (doc1)")

	// 4. 추가 데이터 인덱싱 후 삭제 예제
	doc2 := Document{
		Title:   "업데이트됨",
		Content: "업데이트됨",
	}
	err = index.Index("doc2", doc2)
	if err != nil {
		log.Fatal("두 번째 문서 인덱싱 에러:", err)
	}
	fmt.Println("문서 추가 완료 (doc2)")

	// 문서 삭제 (문서 ID "doc2" 삭제)
	// err = index.Delete("doc2")
	// if err != nil {
	// 	log.Fatal("문서 삭제 에러:", err)
	// }
	// fmt.Println("문서 삭제 완료 (doc2)")

	// 5. 인덱스 저장
	// Bleve는 인덱스 변경 시 자동으로 디스크에 기록하지만,
	// 인덱스를 사용하지 않을 경우에는 Close()를 호출하여 안전하게 닫아줍니다.
	err = index.Close()
	if err != nil {
		log.Fatal("인덱스 닫기 에러:", err)
	}
	fmt.Println("인덱스 저장 및 닫기 완료")

	// 6. 인덱스 다시 로드
	// 디스크에 저장된 인덱스를 다시 열어서 검색 등 작업을 수행할 수 있습니다.
	index, err = bleve.Open(indexPath)
	if err != nil {
		log.Fatal("인덱스 재로드 에러:", err)
	}
	fmt.Println("인덱스 재로드 완료")

	// 7. 검색
	// 예를 들어, "업데이트됨"이라는 단어가 포함된 문서를 검색해봅니다.
	query := bleve.NewMatchQuery("업데이트됨")
	searchRequest := bleve.NewSearchRequest(query)

	searchResult, err := index.Search(searchRequest)
	if err != nil {
		log.Fatal("검색 에러:", err)
	}
	fmt.Println("검색 결과:")
	fmt.Printf("%+v\n", searchResult)
	fmt.Println(searchResult.Hits[0].Index, searchResult.Hits[0].Fields, searchResult.Hits[0].Score)
	// 작업이 끝난 후 인덱스를 닫아줍니다.
	err = index.Close()
	if err != nil {
		log.Fatal("인덱스 최종 닫기 에러:", err)
	}
}
