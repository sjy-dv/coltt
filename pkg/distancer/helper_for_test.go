package distancer

import (
	"math/rand"
	"time"
)

func getRandomSeed() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
