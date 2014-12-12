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

type SetAttrMask int32

const (
	SET_ATTR_MODE = SetAttrMask(1 << iota)
	SET_ATTR_UID
	SET_ATTR_GID
	SET_ATTR_SIZE
	SET_ATTR_ATIME
	SET_ATTR_MTIME
	SET_ATTR_UNUSED_ // placeholder for 1 << 6
	SET_ATTR_ATIME_NOW
	SET_ATTR_MTIME_NOW
)
