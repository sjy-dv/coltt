package compresshelper_test

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/sjy-dv/nnv/pkg/compresshelper"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/stretchr/testify/assert"
)

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

func TestF16Losses128dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(128)
		compareVec := generateRandomVector(128)
		cfloat := make([]compresshelper.Float16, len(randVec))
		ccfloat := make([]compresshelper.Float16, len(compareVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.Fromfloat32(x)
			ccfloat[i] = compresshelper.Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 128)
		recoverFloat2 := make([]float32, 128)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F16 128 dim compress is passed")
}

func TestF8Losses128dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(128)
		compareVec := generateRandomVector(128)
		cfloat := make([]compresshelper.Float8, len(randVec))
		ccfloat := make([]compresshelper.Float8, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.F8Fromfloat32(x)
			ccfloat[i] = compresshelper.F8Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 128)
		recoverFloat2 := make([]float32, 128)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F8 128 dim compress is passed")
}

func TestBF16Losses128dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(128)
		compareVec := generateRandomVector(128)
		cfloat := make([]compresshelper.BFloat16, len(randVec))
		ccfloat := make([]compresshelper.BFloat16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.BF16Fromfloat32(x)
			ccfloat[i] = compresshelper.BF16Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 128)
		recoverFloat2 := make([]float32, 128)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("BF16 128 dim compress is passed")
}

func TestF16Losses384dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(384)
		compareVec := generateRandomVector(384)
		cfloat := make([]compresshelper.Float16, len(randVec))
		ccfloat := make([]compresshelper.Float16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.Fromfloat32(x)
			ccfloat[i] = compresshelper.Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 384)
		recoverFloat2 := make([]float32, 384)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F16 384 dim compress is passed")
}

func TestF8Losses384dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(384)
		compareVec := generateRandomVector(384)
		cfloat := make([]compresshelper.Float8, len(randVec))
		ccfloat := make([]compresshelper.Float8, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.F8Fromfloat32(x)
			ccfloat[i] = compresshelper.F8Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 384)
		recoverFloat2 := make([]float32, 384)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = cfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F8 384 dim compress is passed")
}

func TestBF16Losses384dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(384)
		compareVec := generateRandomVector(384)
		cfloat := make([]compresshelper.BFloat16, len(randVec))
		ccfloat := make([]compresshelper.BFloat16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.BF16Fromfloat32(x)
			ccfloat[i] = compresshelper.BF16Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 384)
		recoverFloat2 := make([]float32, 384)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("BF16 384 dim compress is passed")
}

func TestF16Losses768dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(768)
		compareVec := generateRandomVector(768)
		cfloat := make([]compresshelper.Float16, len(randVec))
		ccfloat := make([]compresshelper.Float16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.Fromfloat32(x)
			ccfloat[i] = compresshelper.Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 768)
		recoverFloat2 := make([]float32, 768)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F16 768 dim compress is passed")
}

func TestF8Losses768dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(768)
		compareVec := generateRandomVector(768)
		cfloat := make([]compresshelper.Float8, len(randVec))
		ccfloat := make([]compresshelper.Float8, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.F8Fromfloat32(x)
			ccfloat[i] = compresshelper.F8Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 768)
		recoverFloat2 := make([]float32, 768)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F8 768 dim compress is passed")
}

func TestBF16Losses768dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(768)
		compareVec := generateRandomVector(768)
		cfloat := make([]compresshelper.BFloat16, len(randVec))
		ccfloat := make([]compresshelper.BFloat16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.BF16Fromfloat32(x)
			ccfloat[i] = compresshelper.BF16Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 768)
		recoverFloat2 := make([]float32, 768)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("BF16 768 dim compress is passed")
}

func TestF16Losses1536dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(1536)
		compareVec := generateRandomVector(1536)
		cfloat := make([]compresshelper.Float16, len(randVec))
		ccfloat := make([]compresshelper.Float16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.Fromfloat32(x)
			ccfloat[i] = compresshelper.Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 1536)
		recoverFloat2 := make([]float32, 1536)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F16 1536 dim compress is passed")
}

func TestF8Losses1536dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(1536)
		compareVec := generateRandomVector(1536)
		cfloat := make([]compresshelper.Float8, len(randVec))
		ccfloat := make([]compresshelper.Float8, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.F8Fromfloat32(x)
			ccfloat[i] = compresshelper.F8Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 1536)
		recoverFloat2 := make([]float32, 1536)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F8 1536 dim compress is passed")
}

func TestBF16Losses1536dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(1536)
		compareVec := generateRandomVector(1536)
		cfloat := make([]compresshelper.BFloat16, len(randVec))
		ccfloat := make([]compresshelper.BFloat16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.BF16Fromfloat32(x)
			ccfloat[i] = compresshelper.BF16Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 1536)
		recoverFloat2 := make([]float32, 1536)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("BF16 1536 dim compress is passed")
}

func TestF16Losses3072dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(3072)
		compareVec := generateRandomVector(3072)
		cfloat := make([]compresshelper.Float16, len(randVec))
		ccfloat := make([]compresshelper.Float16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.Fromfloat32(x)
			ccfloat[i] = compresshelper.Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 3072)
		recoverFloat2 := make([]float32, 3072)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F16 3072 dim compress is passed")
}

func TestF8Losses3072dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(3072)
		compareVec := generateRandomVector(3072)
		cfloat := make([]compresshelper.Float8, len(randVec))
		ccfloat := make([]compresshelper.Float8, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.F8Fromfloat32(x)
			ccfloat[i] = compresshelper.F8Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 3072)
		recoverFloat2 := make([]float32, 3072)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("F8 3072 dim compress is passed")
}

func TestBF16Losses3072dim(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		randVec := generateRandomVector(3072)
		compareVec := generateRandomVector(3072)
		cfloat := make([]compresshelper.BFloat16, len(randVec))
		ccfloat := make([]compresshelper.BFloat16, len(randVec))
		for i, x := range randVec {
			cfloat[i] = compresshelper.BF16Fromfloat32(x)
			ccfloat[i] = compresshelper.BF16Fromfloat32(compareVec[i])
		}
		recoverFloat := make([]float32, 3072)
		recoverFloat2 := make([]float32, 3072)
		for i := range cfloat {
			recoverFloat[i] = cfloat[i].Float32()
			recoverFloat2[i] = ccfloat[i].Float32()
		}
		dist1 := ((distance.NewCosine().Distance(randVec, compareVec) + 1) / 2) * 100
		dist2 := ((distance.NewCosine().Distance(recoverFloat, recoverFloat2) + 1) / 2) * 100
		if math.Abs(float64(dist1-dist2)) > 1 {
			assert.Error(t, errors.New("compress output is failed"))
		}
	}
	t.Log("BF16 3072 dim compress is passed")
}
