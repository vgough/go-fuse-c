package fuse

import (
	"syscall"
)

// Status is the errno number that a FUSE call returns to the kernel.
type Status int32

const (
	OK        = Status(0)
	EACCES    = Status(syscall.EACCES)
	EBUSY     = Status(syscall.EBUSY)
	EEXIST    = Status(syscall.EEXIST)
	EINVAL    = Status(syscall.EINVAL)
	EIO       = Status(syscall.EIO)
	ENOENT    = Status(syscall.ENOENT)
	ENOSYS    = Status(syscall.ENOSYS)
	ENODATA   = Status(syscall.ENODATA)
	ENOTDIR   = Status(syscall.ENOTDIR)
	ENOTEMPTY = Status(syscall.ENOTEMPTY)
	EPERM     = Status(syscall.EPERM)
	ERANGE    = Status(syscall.ERANGE)
	EXDEV     = Status(syscall.EXDEV)
	EBADF     = Status(syscall.EBADF)
	ENODEV    = Status(syscall.ENODEV)
	EROFS     = Status(syscall.EROFS)
	EISDIR    = Status(syscall.EISDIR)
)

type AccessMode int32

const (
	O_RDONLY = AccessMode(0)
	O_WRONLY = AccessMode(1)
	O_RDWR   = AccessMode(2)
)

type FsFlags int32

const (
	ST_RDONLY = FsFlags(1) // Mount read-only
	ST_NOSUID = FsFlags(2) // Ignore suid and sgid bits
)
