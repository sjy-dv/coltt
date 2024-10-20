package flat

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

type IndexFlat struct {
	vecStore vectorspace.VectorStore
}

func NewIndexFlat(params models.IndexVectorParameters, storage storage.Storage) (inf IndexFlat, err error) {
	// ---------------------------
	vstore, err := vectorspace.New(params.Quantizer, storage, params.DistanceMetric, int(params.VectorSize))
	if err != nil {
		err = fmt.Errorf("failed to create vector store: %w", err)
		return
	}
	inf.vecStore = vstore
	// ---------------------------
	return
}

func (inf IndexFlat) SizeInMemory() int64 {
	return inf.vecStore.SizeInMemory()
}

func (inf IndexFlat) UpdateStorage(storage storage.Storage) {
	inf.vecStore.UpdateStorage(storage)
}

func (inf IndexFlat) InsertUpdateDelete(ctx context.Context, points <-chan models.IndexVectorChange) <-chan error {
	sinkErrC := withcontext.SinkWithContext(ctx, points, func(point models.IndexVectorChange) error {
		// Does this point exist?
		var err error
		switch {
		case point.Vector != nil:
			// Insert or update
			_, err = inf.vecStore.Set(point.Id, point.Vector)
		case point.Vector == nil:
			// Delete
			err = inf.vecStore.Delete(point.Id)
		default:
			err = fmt.Errorf("unknown operation for point: %d", point.Id)
		}
		return err
	})
	errC := make(chan error, 1)
	// We use this go routine to flush the vector store after all points have
	// been processed
	go func() {
		defer close(errC)
		if err := <-sinkErrC; err != nil {
			errC <- fmt.Errorf("failed to insert/update/delete: %w", err)
			return
		}
		// Check vector store optimisation
		if err := inf.vecStore.Fit(); err != nil {
			errC <- fmt.Errorf("failed to fit vector store: %w", err)
			return
		}
		errC <- inf.vecStore.Flush()
	}()
	return errC
}

func (inf IndexFlat) Search(ctx context.Context, options models.SearchVectorFlatOptions, filter *roaring64.Bitmap) (*roaring64.Bitmap, []models.SearchResult, error) {
	distFn := inf.vecStore.DistanceFromFloat(options.Vector)
	// ---------------------------
	var weight float32 = 1
	if options.Weight != nil {
		weight = *options.Weight
	}
	// ---------------------------
	/* We used to use multiple workers to scan through the vector store, but for
	 * individual requests coming it adds too much overhead and the gain a low,
	 * around 10 queries per second. So to not suffocate the CPU across requests
	 * we keep it single-threaded. Also no reasonably sized collection should
	 * use flat index as the main one. */
	startTime := time.Now()
	res := make([]models.SearchResult, 0, options.Limit)
	err := inf.vecStore.ForEach(func(point vectorspace.VectorStorePoint) error {
		if filter != nil && !filter.Contains(point.Id()) {
			return nil
		}
		dist := distFn(point)
		// cap here is capacity of the array = options.Limit above, in case you
		// are new to the Go language.
		// Is it worth adding?
		if len(res) == cap(res) && dist >= *res[len(res)-1].Distance {
			return nil
		}
		/* Insert using insertion sort, we don't expect limit (K) to be very
		 * large. We add the element to the end and swap until it is in the right
		 * place. */
		sr := models.SearchResult{
			NodeId:   point.Id(),
			Distance: &dist,
			// We -1 multiply so that the sort order is correct, lower distance
			// higher score
			HybridScore: (-1 * weight * dist),
		}
		if len(res) < cap(res) {
			res = append(res, sr)
		} else {
			res[len(res)-1] = sr
		}
		for i := len(res) - 1; i > 0 && *res[i].Distance < *res[i-1].Distance; i-- {
			res[i], res[i-1] = res[i-1], res[i]
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to iterate over points: %w", err)
	}
	log.Debug().Dur("elapsed", time.Since(startTime)).Msg("search flat")
	// ---------------------------
	rSet := roaring64.New()
	for _, r := range res {
		rSet.Add(r.NodeId)
	}
	return rSet, res, nil
}
