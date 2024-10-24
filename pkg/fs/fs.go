package fs

import "io"

type File interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.Closer
	Truncate(int64) error
	Size() int64
	Sync() error
}

type FileSystem = byte

const (
	OSFileSystem FileSystem = iota
)

func Open(name string, fs FileSystem) (File, error) {
	switch fs {
	case OSFileSystem:
		return openOSFile(name)
	}
	return nil, nil
}
