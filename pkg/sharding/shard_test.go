package sharding

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestShardTraffic(t *testing.T) {
	var c uint64 = 3
	res := make(map[uint64]int)
	for i := 0; i < 999; i++ {
		r := ShardTraffic(uuid.New(), c)
		res[r] += 1
	}
	fmt.Println(res)
	t.Log(res)
}
