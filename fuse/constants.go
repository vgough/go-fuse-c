package fuse

import (
	"syscall"
)

const (
	FUSE_ROOT_ID     = 1 // Inode number of the root node.
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
	EACCES    = Status(syscall.EACCES)    // 13 - permission denied
	EBUSY     = Status(syscall.EBUSY)     // 16 - resource busy
	EEXIST    = Status(syscall.EEXIST)    // 17 - node exists
	EINVAL    = Status(syscall.EINVAL)    // 22 - invalid argument
	EIO       = Status(syscall.EIO)       // 5 - input / output error
	ENOENT    = Status(syscall.ENOENT)    // 2 - no such entry
	ENOSYS    = Status(syscall.ENOSYS)    // 78 - operation not implemented
	ENODATA   = Status(syscall.ENODATA)   // 96 - no data available
	ENOTDIR   = Status(syscall.ENOTDIR)   // 20 - not a directory
	ENOTEMPTY = Status(syscall.ENOTEMPTY) // 39 - directory not empty
	EPERM     = Status(syscall.EPERM)     // 1 - operation not permitted
	ERANGE    = Status(syscall.ERANGE)    // 34 - result not representable
	EXDEV     = Status(syscall.EXDEV)     // 18 - cross-device link
	EBADF     = Status(syscall.EBADF)     // 9 - bad file number
	ENODEV    = Status(syscall.ENODEV)    // 19 - no such device
	EROFS     = Status(syscall.EROFS)     // 30 - read-only file system
	EISDIR    = Status(syscall.EISDIR)    // 21 - is a directory
)

// AccessMode holds flags indicating read or write requirements for Open calls.
type AccessMode int32

const (
	O_RDONLY = AccessMode(0)
	O_WRONLY = AccessMode(1)
	O_RDWR   = AccessMode(2)
)

// FsFlags holds filesystem configuration flags.
type FsFlags int32

const (
	ST_RDONLY = FsFlags(1) // Mount read-only
	ST_NOSUID = FsFlags(2) // Ignore suid and sgid bits
)

// SetAttrMask holds flags indicating which metadata to set in a SetAttr call.
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
