package wal

import "os"

type Options struct {
	DirPath        string
	SegmentSize    int64
	SegmentFileExt string
	BlockCache     uint32
	Sync           bool
	BytesPerSync   uint32
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
	BlockCache:     32 * KB * 10,
	Sync:           false,
	BytesPerSync:   0,
}
