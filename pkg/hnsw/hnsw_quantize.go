package hnsw

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/pkg/withcontext"
	"github.com/sjy-dv/nnv/storage"
)

// IndexHNSW는 HNSW 기반의 인덱스를 나타냅니다.
type IndexHNSW struct {
	hnswIndex *HNSW
	mu        sync.RWMutex
}

// NewIndexHNSW는 새로운 HNSW 인덱스를 초기화합니다.
func NewIndexHNSW(params models.IndexVectorHnswParameters, storage storage.Storage) (IndexHNSW, error) {
	// HNSW 인덱스 초기화: M=4, efConstruction=10, 차원=params.VectorSize, 유클리드 유사도 사용
	hnswIndex, err := NewHNSW(int(params.M), int(params.EfConstruction), int(params.VectorSize), Euclidean)
	if err != nil {
		return IndexHNSW{}, fmt.Errorf("failed to create HNSW index: %w", err)
	}
	return IndexHNSW{
		hnswIndex: hnswIndex,
	}, nil
}

// SizeInMemory는 인덱스의 메모리 사용량을 반환합니다.
func (inf IndexHNSW) SizeInMemory() int64 {
	inf.mu.RLock()
	defer inf.mu.RUnlock()
	return inf.hnswIndex.SizeInMemory()
}

// UpdateStorage는 인덱스의 저장소를 업데이트합니다.
func (inf IndexHNSW) UpdateStorage(storage storage.Storage) {
	inf.mu.Lock()
	defer inf.mu.Unlock()
	inf.hnswIndex.UpdateStorage(storage)
}

// InsertUpdateDelete는 포인트를 삽입, 업데이트 또는 삭제합니다.
func (inf IndexHNSW) InsertUpdateDelete(ctx context.Context, points <-chan models.IndexVectorChange) <-chan error {
	sinkErrC := withcontext.SinkWithContext(ctx, points, func(point models.IndexVectorChange) error {
		inf.mu.Lock()
		defer inf.mu.Unlock()

		switch {
		case point.Vector != nil:
			// 삽입 또는 업데이트
			_, err := inf.hnswIndex.AddPoint(point.Vector)
			return err
		case point.Vector == nil:
			// 삭제
			return inf.hnswIndex.DeletePoint(point.Id)
		default:
			return fmt.Errorf("unknown operation for point: %d", point.Id)
		}
	})
	errC := make(chan error, 1)
	// 모든 포인트가 처리된 후 인덱스를 플러시
	go func() {
		defer close(errC)
		if err := <-sinkErrC; err != nil {
			errC <- fmt.Errorf("failed to insert/update/delete: %w", err)
			return
		}
		if err := inf.hnswIndex.Fit(); err != nil {
			errC <- fmt.Errorf("failed to fit HNSW index: %w", err)
			return
		}
		errC <- inf.hnswIndex.Flush()
	}()
	return errC
}

// Search는 쿼리 벡터에 대한 검색을 수행합니다.
func (inf IndexHNSW) Search(ctx context.Context, options models.SearchVectorFlatOptions, filter *roaring64.Bitmap) (*roaring64.Bitmap, []models.SearchResult, error) {
	inf.mu.RLock()
	defer inf.mu.RUnlock()

	query := options.Vector
	k := options.Limit
	weight := float32(1)
	if options.Weight != nil {
		weight = *options.Weight
	}

	startTime := time.Now()
	results, err := inf.hnswIndex.SearchKNN(query, k, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}
	log.Debug().Dur("elapsed", time.Since(startTime)).Msg("search HNSW")

	// 결과를 RoaringBitmap 및 SearchResult 슬라이스로 변환
	rSet := roaring64.New()
	searchResults := make([]models.SearchResult, 0, len(results))
	for _, res := range results {
		rSet.Add(res.ID)
		searchResults = append(searchResults, models.SearchResult{
			NodeId:      res.ID,
			Distance:    &res.Score,
			HybridScore: (-1 * weight * res.Score),
		})
	}
	return rSet, searchResults, nil
}
