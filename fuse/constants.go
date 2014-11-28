package fuse

import (
	"syscall"
)

const (
	FUSE_ROOT_ID     = 1
	FUSE_UNKNOWN_INO = 0xffffffff

	S_IFDIR = syscall.S_IFDIR
	S_IFREG = syscall.S_IFREG
	S_IFLNK = syscall.S_IFLNK
	S_IFIFO = syscall.S_IFIFO
)
