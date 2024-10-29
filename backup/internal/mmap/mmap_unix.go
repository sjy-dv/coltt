//go:build !windows && !plan9
// +build !windows,!plan9

package mmap

import "syscall"

func mmap(fd, length int) ([]byte, error) {
	return syscall.Mmap(
		fd,
		0,
		length,
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
}
