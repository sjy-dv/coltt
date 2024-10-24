package fs

import (
	"os"
)

type OSFile struct {
	fd   *os.File
	size int64
}

func openOSFile(name string) (File, error) {
	fd, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// get the file size
	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}
	return &OSFile{fd: fd, size: stat.Size()}, nil
}

func (of *OSFile) Read(p []byte) (n int, err error) {
	return of.fd.Read(p)
}

func (of *OSFile) ReadAt(b []byte, off int64) (n int, err error) {
	return of.fd.ReadAt(b, off)
}

func (of *OSFile) Write(p []byte) (n int, err error) {
	return of.fd.Write(p)
}

func (of *OSFile) WriteAt(b []byte, off int64) (n int, err error) {
	return of.fd.WriteAt(b, off)
}

func (of *OSFile) Truncate(size int64) error {
	err := of.fd.Truncate(of.size + size)
	if err != nil {
		return err
	}
	of.size += size
	return nil
}

func (of *OSFile) Size() int64 {
	return of.size
}

func (of *OSFile) Sync() error {
	return of.fd.Sync()
}

func (of *OSFile) Close() error {
	return of.fd.Close()
}
