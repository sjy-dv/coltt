package sharding

import (
	"encoding/binary"

	"github.com/google/uuid"
)

// c values .. server count or data shard count
func ShardTraffic(x uuid.UUID, c uint64) uint64 {
	res := ((binary.LittleEndian.Uint64(x[:8]) % c) +
		(binary.BigEndian.Uint64(x[8:]) % c))
	return res % c
}
