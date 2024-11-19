package wal

import "os"

type Options struct {
	DirPath string

	SegmentSize int64

	SegmentFileExt string

	Sync bool

	BytesPerSync uint32
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

var DefaultOptions = Options{
	DirPath:        os.TempDir(),
	SegmentSize:    GB,
	SegmentFileExt: ".SEG",
	Sync:           false,
	BytesPerSync:   0,
}
