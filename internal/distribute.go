package internal

import (
	"encoding/binary"

	"github.com/google/uuid"
)

func DistributeUuid(x uuid.UUID, partition uint64) uint64 {
	res := ((binary.LittleEndian.Uint64(x[:8]) % partition) +
		(binary.LittleEndian.Uint64(x[8:]) % partition))
	return res % partition
}
