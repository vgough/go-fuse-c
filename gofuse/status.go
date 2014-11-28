package gofuse

import (
	"syscall"
)

// Status is the errno number that a FUSE call returns to the kernel.
type Status int32

const (
	OK      = Status(0)
	EACCES  = Status(syscall.EACCES)
	EBUSY   = Status(syscall.EBUSY)
	EINVAL  = Status(syscall.EINVAL)
	EIO     = Status(syscall.EIO)
	ENOENT  = Status(syscall.ENOENT)
	ENOSYS  = Status(syscall.ENOSYS)
	ENODATA = Status(syscall.ENODATA)
	ENOTDIR = Status(syscall.ENOTDIR)
	EPERM   = Status(syscall.EPERM)
	ERANGE  = Status(syscall.ERANGE)
	EXDEV   = Status(syscall.EXDEV)
	EBADF   = Status(syscall.EBADF)
	ENODEV  = Status(syscall.ENODEV)
	EROFS   = Status(syscall.EROFS)
)
