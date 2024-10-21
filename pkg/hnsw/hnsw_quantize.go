package hnsw

import (
	"context"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/pkg/vectorspace"
	"github.com/sjy-dv/nnv/pkg/withcontext"
	"github.com/sjy-dv/nnv/storage"
)

type IndexHNSW struct {
	hnswIndex *HNSW
	vecStore  vectorspace.VectorStore
}

func NewIndexHNSW(params models.IndexVectorHnswParameters, storage storage.Storage) (inh IndexHNSW, err error) {
	hnswIndex, err := NewHNSW(int(params.M), int(params.EfConstruction), int(params.VectorSize), Euclidean)
	if err != nil {
		return IndexHNSW{}, fmt.Errorf("failed to create HNSW index: %w", err)
	}
	vstore, err := vectorspace.New(params.Quantizer, storage, params.DistanceMetric, int(params.VectorSize))
	if err != nil {
		err = fmt.Errorf("failed to create vector store: %w", err)
		return
	}
	return IndexHNSW{
		hnswIndex: hnswIndex,
		vecStore:  vstore,
	}, nil
}

func (inf IndexHNSW) SizeInMemory() int64 {
	return (inf.hnswIndex.SizeInMemory() + inf.vecStore.SizeInMemory())
}

func (inf IndexHNSW) UpdateStorage(storage storage.Storage) {
	inf.hnswIndex.UpdateStorage(storage)
	inf.vecStore.UpdateStorage(storage)
}

func (inf IndexHNSW) InsertUpdateDelete(ctx context.Context, points <-chan models.IndexVectorChange) <-chan error {
	sinkErrC := withcontext.SinkWithContext(ctx, points, func(point models.IndexVectorChange) error {

		switch {
		case point.Vector != nil:

			_, err := inf.hnswIndex.AddPoint(point.Vector, point.Id)
			if err != nil {
				return err
			}
			_, err = inf.vecStore.Set(point.Id, point.Vector)
			return err
		case point.Vector == nil:

			err := inf.hnswIndex.DeletePoint(point.Id)
			if err != nil {
				return err
			}
			err = inf.vecStore.Delete(point.Id)
			return err
		default:
			return fmt.Errorf("unknown operation for point: %d", point.Id)
		}
	})
	errC := make(chan error, 1)
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

func (inf IndexHNSW) Search(ctx context.Context, options models.SearchVectorFlatOptions, filter *roaring64.Bitmap) (*roaring64.Bitmap, []models.SearchResult, error) {

	query := options.Vector
	k := options.Limit
	weight := float32(1)
	if options.Weight != nil {
		weight = *options.Weight
	}

	startTime := time.Now()
	results, err := inf.hnswIndex.Search(query, k, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}
	log.Debug().Dur("elapsed", time.Since(startTime)).Msg("search HNSW")

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
