package hnsw_test

import (
	"cmp"
	"context"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"testing"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/rs/zerolog"
	"github.com/sjy-dv/nnv/pkg/conversion"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/pkg/withcontext"
	"github.com/sjy-dv/nnv/storage"
	"github.com/stretchr/testify/require"
)

var hnswParams = models.IndexVectorHnswParameters{
	VectorSize:     2,
	DistanceMetric: "euclidean",
	M:              16,
	EfConstruction: 100,
}

// recommend params
// if < 100_000
// M: 16 ef: 100
// if < 1_000_000
// M: 32 ef: 200
// if < 10_000_000
// M: 48 ef: 300
// high level dimension
// M: 32 ef: 200
func checkVectorCount(t *testing.T, storage storage.Storage, expected int) {
	t.Helper()
	// Check vector count

	count := 0
	err := storage.ForEach(func(key []byte, value []byte) error {
		_, ok := conversion.NodeIdFromKey(key, 'v')
		if ok {
			count++
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, expected, count)
}
func randPoints(size, offset int) []models.IndexVectorChange {
	points := make([]models.IndexVectorChange, size)
	vectorSize := 2
	for i := 0; i < size; i++ {
		randVector := make([]float32, vectorSize)
		sum := float32(0)
		for j := 0; j < vectorSize; j++ {
			randVector[j] = rand.Float32()
			sum += randVector[j]
		}
		for j := 0; j < vectorSize; j++ {
			randVector[j] /= sum
		}
		points[i] = models.IndexVectorChange{
			// 0 is not allowed, 1 is start node
			Id:     uint64(i + offset + 2),
			Vector: randVector,
		}
	}
	return points
}

func Test_ConcurrentCUD(t *testing.T) {
	storj := storage.NewMemStorage(false)
	inv, err := hnsw.NewIndexHNSW(hnswParams, storj)
	require.NoError(t, err)
	// Pre-insert
	in := make(chan models.IndexVectorChange)
	errC := inv.InsertUpdateDelete(context.Background(), in)
	for _, rp := range randPoints(50, 0) {
		in <- rp
	}
	// ---------------------------
	var wg sync.WaitGroup
	wg.Add(3)
	// Insert more
	go func() {
		for _, rp := range randPoints(50, 50) {
			in <- rp
		}
		wg.Done()
	}()
	// ---------------------------
	// Update some
	go func() {
		for _, rp := range randPoints(25, 25) {
			in <- rp
		}
		wg.Done()
	}()
	// ---------------------------
	// Delete some
	go func() {
		for i := 0; i < 25; i++ {
			in <- models.IndexVectorChange{Id: uint64(i + 2), Vector: nil}
		}
		wg.Done()
	}()
	// ---------------------------
	wg.Wait()
	close(in)
	require.NoError(t, <-errC)
	checkVectorCount(t, storj, 75)
}

func Test_Search(t *testing.T) {
	bucket := storage.NewMemStorage(false)
	inv, err := hnsw.NewIndexHNSW(hnswParams, bucket)
	require.NoError(t, err)
	// Pre-insert
	ctx := context.Background()
	rps := randPoints(50, 0)
	in := withcontext.ProduceWithContext(ctx, rps)
	errC := inv.InsertUpdateDelete(ctx, in)
	require.NoError(t, <-errC)
	// ---------------------------
	// Search
	options := models.SearchVectorFlatOptions{
		Vector: rps[0].Vector,
		Limit:  10,
	}
	filter := roaring64.BitmapOf(rps[0].Id)
	rSet, results, err := inv.Search(ctx, options, filter)
	require.NoError(t, err)
	require.EqualValues(t, 1, rSet.GetCardinality())
	require.Len(t, results, 1)
	require.Equal(t, rps[0].Id, results[0].NodeId)
	require.Equal(t, float32(0), *results[0].Distance)
}

func Test_Recall(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	distFnNames := []string{models.DistanceEuclidean}
	for _, distFnName := range distFnNames {
		testName := fmt.Sprintf("distFn=%s", distFnName)
		t.Run(testName, func(t *testing.T) {
			bucket := storage.NewMemStorage(false)
			params := models.IndexVectorHnswParameters{
				VectorSize:     2,
				DistanceMetric: distFnName,
				M:              16,
				EfConstruction: 100,
			}
			inv, err := hnsw.NewIndexHNSW(params, bucket)
			require.NoError(t, err)
			// Pre-insert
			ctx := context.Background()
			rps := randPoints(2000, 0)
			in := withcontext.ProduceWithContext(ctx, rps)
			errC := inv.InsertUpdateDelete(ctx, in)
			require.NoError(t, <-errC)
			// ---------------------------
			// Search
			options := models.SearchVectorFlatOptions{
				Vector: rps[0].Vector,
				Limit:  10,
			}
			// ---------------------------
			// Find ground truth
			groundTruth := make([]models.SearchResult, 0)
			for _, rp := range rps {
				distFn, _ := distance.GetFloatDistanceFn(params.DistanceMetric)
				dist := distFn(options.Vector, rp.Vector)
				groundTruth = append(groundTruth, models.SearchResult{
					NodeId:   rp.Id,
					Distance: &dist,
				})
			}
			slices.SortFunc(groundTruth, func(a, b models.SearchResult) int {
				return cmp.Compare(*a.Distance, *b.Distance)
			})
			groundTruth = groundTruth[:options.Limit]
			// ---------------------------
			rSet, results, err := inv.Search(ctx, options, nil)
			fmt.Println(rSet, results)
			require.NoError(t, err)
			require.EqualValues(t, 10, rSet.GetCardinality())
			require.Len(t, results, 10)
			for i, res := range results {
				require.Equal(t, *groundTruth[i].Distance, *res.Distance)
				// The ordering might not be exact if the distances are the
				// same, maybe one after or one before
				if groundTruth[i].NodeId != res.NodeId && i > 0 && i < len(results)-1 {
					nextId := results[i+1].NodeId
					prevId := results[i-1].NodeId
					require.True(t, groundTruth[i].NodeId == nextId || groundTruth[i].NodeId == prevId)
				}
			}
		})
	}
}
