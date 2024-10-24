package diskhash

import "os"

type Options struct {
	DirPath         string
	SlotValueLength uint32
	LoadFactor      float64
}

var DefaultOptions = Options{
	DirPath:         os.TempDir(),
	SlotValueLength: 0,
	LoadFactor:      0.7,
}
