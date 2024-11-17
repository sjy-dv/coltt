package distancepq

import (
	"math"
	"math/bits"

	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/distance"
)

type FloatDistFunc func(x, y []float32) float32
type BitDistFunc func(x, y []uint64) float32

// Euclidean distance actually computes the squared euclidean distance for efficiency. This should not affect the
// results of the nearest neighbour search as the square root is monotonic.
var euclideanDistance FloatDistFunc = squaredEuclideanDistancePureGo
var dotProductImpl FloatDistFunc = dotProductPureGo

func dotProductDistance(x, y []float32) float32 {
	return -dotProductImpl(x, y)
}

func cosineDistance(x, y []float32) float32 {
	return 1 - dotProductImpl(x, y)
}

const degToRad = math.Pi / 180

// Earth radius in meters
const earthRadius = 6371000

// Computes the haversine distance between two points on the Earth's surface. It
// assumes [lat, long] coordinates in degrees.
// Formula credit: https://scikit-learn.org/stable/modules/generated/sklearn.metrics.pairwise.haversine_distances.html
func haversineDistance(x, y []float32) float32 {
	latx, lonx, laty, lony := float64(x[0])*degToRad, float64(x[1])*degToRad, float64(y[0])*degToRad, float64(y[1])*degToRad
	dlat, dlon := latx-laty, lonx-lony
	// Please see the formula in the link above for more details.
	sinDlat, sinDlon := math.Sin(dlat/2), math.Sin(dlon/2)
	a := sinDlat*sinDlat + math.Cos(latx)*math.Cos(laty)*sinDlon*sinDlon
	c := 2 * math.Asin(math.Sqrt(a))
	return float32(earthRadius * c)
}

func hammingDistance(x, y []uint64) float32 {
	dist := 0
	for i := range x {
		// The XOR ^ operator returns a 1 in each bit position for which the
		// corresponding bits of the two operands are different. Then we count
		// the number of bits that are different.
		dist += bits.OnesCount64(x[i] ^ y[i])
	}
	return float32(dist)
}

func jaccardDistance(x, y []uint64) float32 {
	intersection := 0
	union := 0
	for i := range x {
		intersection += bits.OnesCount64(x[i] & y[i])
		union += bits.OnesCount64(x[i] | y[i])
	}
	if union == 0 {
		return 0
	}
	return 1 - float32(intersection)/float32(union)
}

// Returns floating distance function by name.
func GetFloatDistanceFn(name string) FloatDistFunc {
	switch name {
	case edge.EUCLIDEAN:
		return distance.NewEuclidean().Distance
	case edge.COSINE:
		return distance.NewCosine().Distance
	default:
		return distance.NewEuclidean().Distance
	}
}

// func GetBitDistanceFn(name string) (BitDistFunc, error) {
// 	switch name {
// 	case models.DistanceHamming:
// 		return hammingDistance, nil
// 	case models.DistanceJaccard:
// 		return jaccardDistance, nil
// 	default:
// 		return nil, fmt.Errorf("unknown bit distance function: %s", name)
// 	}
// }
