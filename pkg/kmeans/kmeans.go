package kmeans

import (
	"math"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/models"
)

/* This kmeans is geared towards clustering subvectors which originated from
 * product quantization. It follows the standard naive algorithm for KMeans, also
 * known as Lloyd's algorithm. */
type KMeans struct {
	// Number of clusters, maximum 256 because we use uint8 for labels to limit
	// memory usage
	K int
	// Number of iterations
	MaxIter int
	// Offset into features array, this is geared towards clustering subvectors
	Offset int
	// Vector length, used as x[offset:offset+vectorLen]
	VectorLen int
	// ---------------------------
	// Cluster centroids
	Centroids [][]float32
	// Labels
	Labels []uint8
}

// Fit performs the KMeans clustering algorithm on the given data.
func (km *KMeans) Fit(X [][]float32) {
	/* We will perform the standard naive algorithm for KMeans, also known as
	 * Lloyd's algorithm. It consists of two stages: an assignment stage and an
	 * update stage.
	 *
	 * During the assignment stage, we assign each sample to the nearest cluster
	 * center. Then during the update stage, we update the cluster centers to be
	 * the mean of the samples assigned to them.
	 *
	 * Read more at: https://en.wikipedia.org/wiki/K-means_clustering
	 */
	distFn, _ := distance.GetFloatDistanceFn(models.DistanceEuclidean)
	logger := log.With().Str("module", "kmeans").Int("size", len(X)).Logger()
	// ---------------------------
	/* Initialise centroids following kmeans++, we effectively find the furthest
	 * points from existing centroids. Although this obviously takes more time,
	 * it actually helps converge faster leading to better clustering. We also
	 * don't expect to be performing clustering on large amounts of data. */
	// Keeps tracks of the distance to nearest centroid
	startTime := time.Now()
	centroidDists := make([]float32, len(X))
	for i := 0; i < len(X); i++ {
		centroidDists[i] = math.MaxFloat32
	}
	// Assign one random point as the first centroid
	alreadyCentroid := make(map[int]struct{})
	km.Centroids = make([][]float32, km.K)
	randId := rand.Intn(len(X))
	alreadyCentroid[randId] = struct{}{}
	km.Centroids[0] = X[randId][km.Offset : km.Offset+km.VectorLen]
	// For the remainder find the furthest point from the existing centroids
	for i := 1; i < km.K; i++ {
		furthestDist := float32(0)
		furthestId := 0
		for j := 0; j < len(X); j++ {
			if _, ok := alreadyCentroid[j]; ok {
				continue
			}
			subVec := X[j][km.Offset : km.Offset+km.VectorLen]
			centroidDist := distFn(subVec, km.Centroids[i-1])
			if centroidDist < centroidDists[j] {
				centroidDists[j] = centroidDist
			}
			if centroidDists[j] > furthestDist {
				furthestDist = centroidDists[j]
				furthestId = j
			}
		}
		km.Centroids[i] = X[furthestId][km.Offset : km.Offset+km.VectorLen]
	}
	logger.Debug().Dur("duration", time.Since(startTime)).Msg("initialising centroids")
	// ---------------------------
	km.Labels = make([]uint8, len(X))
	// The sums are used to calculate the mean and then update the centroids
	centroidSums := make([][]float32, km.K)
	for i := 0; i < km.K; i++ {
		centroidSums[i] = make([]float32, km.VectorLen)
	}
	centroidCounts := make([]int, km.K)
	// ---------------------------
	startTime = time.Now()
	for iter := 0; iter < km.MaxIter; iter++ {
		// ---------------------------
		// Assignment stage, answer the question: which cluster does each point
		// belong to?
		changeCount := 0
		for i, x := range X {
			subVec := x[km.Offset : km.Offset+km.VectorLen]
			closestCentroidDist := distFn(subVec, km.Centroids[0])
			closestCentroidId := uint8(0)
			for j := 1; j < km.K; j++ {
				dist := distFn(subVec, km.Centroids[j])
				if dist < closestCentroidDist {
					closestCentroidDist = dist
					closestCentroidId = uint8(j)
				}
			}
			if km.Labels[i] != closestCentroidId {
				changeCount++
				km.Labels[i] = closestCentroidId
			}
		}
		if changeCount == 0 {
			break
		}
		// ---------------------------
		// Update stage, answer the question: what is the new centroid of each cluster?
		for i := 0; i < km.K; i++ {
			centroidCounts[i] = 0
		}
		// Compute sums
		for i, label := range km.Labels {
			// Reset the centroid sum
			if centroidCounts[label] == 0 {
				for j := 0; j < km.VectorLen; j++ {
					centroidSums[label][j] = 0
				}
			}
			centroidCounts[label]++
			subVector := X[i][km.Offset : km.Offset+km.VectorLen]
			for j := 0; j < km.VectorLen; j++ {
				centroidSums[label][j] += subVector[j]
			}
		}
		// Compute mean
		for i := 0; i < km.K; i++ {
			if centroidCounts[i] == 0 {
				continue
			}
			for j := 0; j < km.VectorLen; j++ {
				km.Centroids[i][j] = centroidSums[i][j] / float32(centroidCounts[i])
			}
		}
		// ---------------------------
	}
	logger.Debug().Dur("duration", time.Since(startTime)).Msg("fitting KMeans")
}
