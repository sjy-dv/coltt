package main

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

func main() {
	id := uuid.New()
	fmt.Println(id)
	fmt.Println(UuidMod(id, 3))
}

func UuidMod(x uuid.UUID, mod uint64) uint64 {
	res := ((binary.LittleEndian.Uint64(x[:8]) % mod) + (binary.LittleEndian.Uint64(x[8:]) % mod))
	return res % mod
}
