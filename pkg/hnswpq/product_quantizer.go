package hnswpq

import (
	"errors"
	"math"

	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/distancepq"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

const productQuantizerCentroidDistsKey = "_productQuantizerCentroidDists"
const productQuantizerFlatCentroidsKey = "_productQuantizerFlatCentroids"

type productQuantizer struct {
	params            ProductQuantizerParameters
	distFn            distancepq.FloatDistFunc
	distFnName        string
	originalVectorLen int
	subVectorLen      int

	caches        *ProductQuantizerCache
	centroidDists []float32
	flatCentroids []float32
}

func newProductQuantizer(distFnName string, params ProductQuantizerParameters, vectorLen int) (
	*productQuantizer, error) {

	if vectorLen%params.NumSubVectors != 0 {
		return nil, errors.New("Vector dimensions must be divided equally by the number of subvectors.")
	}

	//     ---

	// # Understanding the Comment on Product Quantization and Cosine Distance

	// **Product Quantization (PQ)** is a technique used to efficiently approximate, store, and search high-dimensional vectors. This method involves splitting each vector into multiple subvectors and independently clustering each subvector to store the cluster centroids. By representing the entire vector as a combination of these cluster indices, PQ reduces memory usage and accelerates search operations.

	// Let's break down the comment you provided step by step:

	// ## 1. Cosine Distance Can't Be Handled Part-Wise

	// - **Cosine Distance** measures the angle between two vectors, focusing on their direction rather than their magnitude.
	// - **Product Quantization** divides each vector into several parts (subvectors) and calculates the distance between each subvector and its corresponding cluster centroid. These distances are then summed to approximate the total distance between the original vectors.
	// - However, cosine distance relies on the overall direction of the entire vector, making it difficult to accurately capture this information when distances are calculated independently for each subvector. Summing these part-wise distances does not effectively represent the cosine distance between the full vectors.

	// ## 2. Product Quantization Splits Each Vector into Parts and Sums Distances to Centroids

	// - PQ works by dividing each high-dimensional vector into smaller, manageable subvectors.
	// - Each subvector is clustered using algorithms like k-means, and the distance between a subvector and its nearest centroid is computed using Euclidean distance.
	// - The total distance between two original vectors is approximated by summing the Euclidean distances of their corresponding subvectors to their respective centroids.

	// ## 3. Even if We Compensate for Cosine, k-Means Clustering Uses Euclidean Distance

	// - **k-Means Clustering** inherently relies on Euclidean distance to determine cluster assignments and update centroids.
	// - Attempting to adjust or compensate for cosine distance within the PQ framework is challenging because the fundamental distance metric used during clustering remains Euclidean.
	// - As a result, the benefits of using cosine distance are not fully realized within the PQ process.

	// ## 4. The Normalization Property of Subvectors Is Lost

	// - **Normalization** involves scaling vectors to have a unit length, which is essential for accurately computing cosine distance.
	// - When PQ processes subvectors independently, the normalization applied to the entire vector may not be preserved for each subvector.
	// - This loss of normalization means that the relationship between Euclidean distance and cosine distance is disrupted, diminishing the effectiveness of cosine-based similarity measures.

	// ## 5. We Still Have Hope Because for Normalized Vectors Euclidean Distance ≈ 2 × Cosine Distance

	// - If vectors are pre-normalized (i.e., scaled to unit length), there exists a proportional relationship between Euclidean distance and cosine distance:

	//   \[
	//   \text{Euclidean Distance} = \sqrt{2 \cdot (1 - \cos(\theta))} \approx 2 \cdot \text{Cosine Distance}
	//   \]

	// - This approximation holds true for normalized vectors, meaning that Euclidean distance becomes proportional to cosine distance.
	// - As a result, even though PQ uses Euclidean distance internally, the proportional relationship allows PQ to effectively approximate cosine distance for normalized vectors, leading to similar search results.

	// ## **In Summary**

	// Product Quantization primarily operates based on Euclidean distance, which poses challenges when trying to use cosine distance directly. Cosine distance relies on the overall direction of vectors, which isn't fully captured when distances are computed for individual subvectors and then summed. However, if vectors are normalized beforehand, the relationship between Euclidean distance and cosine distance becomes proportional. This proportionality allows PQ to approximate cosine distance reasonably well, enabling similar search performance despite the inherent limitations.

	// ---

	if distFnName == edge.COSINE {
		distFnName = edge.EUCLIDEAN
	}

	if params.NumCentroids > 256 {
		return nil, errors.New("There can be no more than 256 centroids.")
	}
	distFn := distancepq.GetFloatDistanceFn(distFnName)

	pq := &productQuantizer{
		params:            params,
		distFn:            distFn,
		distFnName:        distFnName,
		originalVectorLen: vectorLen,
		subVectorLen:      vectorLen / params.NumSubVectors,
	}
	// if alraedy centroid info => load

	return pq, nil
}

func (pq productQuantizer) centroidDistIdx(subvector, centroidX, centroidY int) int {
	return subvector*pq.params.NumCentroids*pq.params.NumCentroids + centroidX*pq.params.NumCentroids + centroidY
}

func (pq productQuantizer) flatCentroidSlice(subvector, centroid int) (start, end int) {
	start = subvector*pq.params.NumCentroids*pq.subVectorLen + centroid*pq.subVectorLen
	end = start + pq.subVectorLen
	return
}

func (pq productQuantizer) encode(vector []float32) []uint8 {
	if len(pq.flatCentroids) == 0 {
		return nil
	}
	/* We will now find the closest centroid for each subvector. */
	encoded := make([]uint8, pq.params.NumSubVectors)
	for i := 0; i < pq.params.NumSubVectors; i++ {
		// The subvector is the slice of the original vector
		subVector := vector[i*pq.subVectorLen : (i+1)*pq.subVectorLen]
		closestCentroidDistance := float32(math.MaxFloat32)
		closestCentroidId := 0
		for j := 0; j < pq.params.NumCentroids; j++ {
			sliceStart, sliceEnd := pq.flatCentroidSlice(i, j)
			centroid := pq.flatCentroids[sliceStart:sliceEnd]
			dist := pq.distFn(subVector, centroid)
			if dist < closestCentroidDistance {
				closestCentroidDistance = dist
				closestCentroidId = j
			}
		}
		encoded[i] = uint8(closestCentroidId)
	}
	return encoded
}

func (pq *productQuantizer) Set(id uint64, vector gomath.Vector) (*productQuantizedPoint, error) {
	point := &productQuantizedPoint{
		id:          id,
		Vector:      vector,
		CentroidIds: pq.encode(vector),
	}
	pq.caches.Put(id, point)
	return point, nil
}

func (pq *productQuantizer) Get(id uint64) (*productQuantizedPoint, error) {
	return pq.caches.Get(id)
}

func (pq *productQuantizer) Delete(ids ...uint64) error {
	return pq.caches.Delete(ids...)
}

func (pq *productQuantizer) Dirty(id uint64) {
	pq.caches.Dirty(id)
	return
}

func (pq *productQuantizer) DistanceFromFloat(x []float32) PointIdDistFn {
	if len(pq.flatCentroids) == 0 {
		return func(y *productQuantizedPoint) float32 {

			return pq.distFn(x, y.Vector)
		}
	}

	dists := make([]float32, pq.params.NumSubVectors*pq.params.NumCentroids)
	for i := 0; i < pq.params.NumSubVectors; i++ {
		subvector := x[i*pq.subVectorLen : (i+1)*pq.subVectorLen]
		for j := 0; j < pq.params.NumCentroids; j++ {
			start, end := pq.flatCentroidSlice(i, j)
			centroid := pq.flatCentroids[start:end]
			dists[i*pq.params.NumCentroids+j] = pq.distFn(subvector, centroid)
		}
	}
	return func(y *productQuantizedPoint) float32 {
		var dist float32
		for i := 0; i < pq.params.NumSubVectors; i++ {
			dist += dists[i*pq.params.NumCentroids+int(y.CentroidIds[i])]
		}
		return dist
	}
}
