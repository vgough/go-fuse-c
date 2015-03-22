package fuse

import (
	"time"
)

// Raw operations for Fuse's LowLevel API.
type RawFileSystem interface {
	// Init initializes a filesystem.
	// Called before any other filesystem method.
	Init(*ConnInfo)

	// Destroy cleans up a filesystem.
	// Called on filesystem exit.
	Destroy()

	// StatFs gets file system statistics.
	StatFs(ino int64) (*StatVfs, Status)

	// Lookup finds a directory entry by name and get its attributes.
	Lookup(dir int64, name string) (*Entry, Status)

	// Forget limits the lifetime of an inode.
	//
	// The n parameter indicates the number of lookups previously performed on this inode.
	// The filesystem may ignore forget calls if the inodes don't need to have a limited lifetime.
	// On unmount it is not guaranteed that all reference dinodes will receive a forget message.
	Forget(ino int64, n int)

	// Release drops an open file reference.
	//
	// Release is called when there are no more references to an open file: all file descriptors are
	// closed and all memory mappings are unmapped.
	//
	// For every open call, there will be exactly one release call.
	//
	// A filesystem may reply with an error, but error values are not returned to the close() or
	// munmap() which triggered the release.
	//
	// fi.Handle will contain the value set by the open method, or will be undefined if the open
	// method didn't set any value.
	// fi.Flags will contain the same flags as for open.
	Release(ino int64, fi *FileInfo) Status

	// Flush is called on each close() of an opened file.
	//
	// Since file descriptors can be duplicated (dup, dup2, fork), for one open call there may be
	// many flush calls.
	//
	// fi.Handle will contain the value set by the open method, or will be undefined if the open
	// method didn't set any value.
	//
	// The name of the method is misleading. Unlike fsync, the filesystem is not forced to flush
	// pending writes.
	Flush(ino int64, fi *FileInfo) Status

	// Fsync synchronizes file contents.
	//
	// If the dataOnly parameter is true, then only the user data should be flushed, not the
	// metdata.
	FSync(ino int64, dataOnly bool, fi *FileInfo) Status

	// Getattr gets file attributes.
	//
	// fi is for future use, currently always nil.
	GetAttr(ino int64, fi *FileInfo) (attr *InoAttr, err Status)

	// Setattr sets file attributes.
	//
	// In the 'attr' argument, only members indicated by the mask contain valid values.  Other
	// members contain undefined values.
	//
	// If the setattr was invoked from the ftruncate() system call, the fi.Handle will contain the
	// value set by the open method.  Otherwise, the fi argument may be nil.
	SetAttr(ino int64, attr *InoAttr, mask SetAttrMask, fi *FileInfo) (*InoAttr, Status)

	// ReadLink reads a symbolic link.
	ReadLink(ino int64) (string, Status)

	// ReadDir reads a directory.
	//
	// fi.Handle will contain the value set by the opendir method, or will be undefined if the
	// opendir method didn't set any value.
	//
	// DirEntryWriter is used to add entries to the output buffer.
	ReadDir(ino int64, fi *FileInfo, off int64, size int, w DirEntryWriter) Status

	// OpenDir opens a directory.
	//
	// Filesystems may store an arbitrary file handle in fh.Handle and use this in other directory
	// operations (ReadDir, ReleaseDir, FsyncDir). Filesystems may not store anything in fi.Handle,
	// though that makes it impossible to implement standard conforming directory stream operations
	// in case the contents of the directory can change between opendir and releasedir.
	OpenDir(ino int64, fi *FileInfo) Status

	// ReleaseDir drops an open file reference.
	//
	// For every OpenDir call, there will be exactly one ReleaseDir call.
	//
	// fi.Handle will contain the value set by the OpenDir method, or will be undefined if the
	// OpenDir method didn't set any value.
	ReleaseDir(ino int64, fi *FileInfo) Status

	// FsyncDir synchronizes directory contents.
	//
	// If the dataOnly parameter is true, then only the user data should be flushed, not the
	// metdata.
	FSyncDir(ino int64, dataOnly bool, fi *FileInfo) Status

	// Mkdir creates a directory.
	Mkdir(parent int64, name string, mode int) (*Entry, Status)

	// Rmdir removes a directory.
	Rmdir(parent int64, name string) Status

	// Rename renames a file or directory.
	Rename(dir int64, name string, newdir int64, newname string) Status

	// Symlink creates a symbolic link.
	Symlink(link string, parent int64, name string) (*Entry, Status)

	// Link creates a hard link.
	Link(ino int64, newparent int64, name string) (*Entry, Status)

	// Mknod creates a file node.
	//
	// This is used to create a regular file, or special files such as character devices, block
	// devices, fifo or socket nodes.
	Mknod(parent int64, name string, mode int, rdev int) (*Entry, Status)

	// Open makes a file available for read or write.
	//
	// Open flags are available in fi.Flags
	//
	// Filesystems may store an arbitrary file handle in fh.Handle and use this in other file
	// operations (read, write, flush, release, fsync). Filesystems may also implement stateless file
	// I/O and not store anything in fi.Handle.
	Open(ino int64, fi *FileInfo) Status

	// Read reads data from an open file.
	//
	// Read should return exactly the number of bytes requested except on EOF or error.
	//
	// fi.Handle will contain the value set by the open method, if any.
	Read(p []byte, ino int64, off int64, fi *FileInfo) (n int, err Status)

	// Write writes data to an open file.
	//
	// Write should return exactly the number of bytes requested except on error.
	//
	// fi.handle will contain the value set by the open method, if any.
	Write(p []byte, ino int64, off int64, fi *FileInfo) (n int, err Status)

	// Unlink removes a file.
	Unlink(parent int64, name string) Status

	// Access checks file access permissions.
	//
	// This will be called for the access() system call.  If the 'default_permissions' mount option
	// is given, this method is not called.
	Access(ino int64, mask int) Status

	// Create creates and opens a file.
	//
	// If the file does not exist, first create it with the specified mode and then open it.
	//
	// Open flags are available in fi.Flags.
	//
	// Filesystems may store an arbitrary file handle in fi.Handle and use this in all other file
	// operations (Read, Write, Flush, Release, FSync).
	//
	// If this method is not implemented, then Mknod and Open methods will be called instead.
	Create(parent int64, name string, mode int, fi *FileInfo) (*Entry, Status)

	// Returns a list of the extended attribute keys.
	ListXattrs(ino int64) ([]string, Status)

	// Returns the size of the attribute value.
	GetXattrSize(ino int64, name string) (int, Status)

	// Get an extended attribute.
	// Result placed in out buffer.
	// Returns the number of bytes copied.
	GetXattr(ino int64, name string, out []byte) (int, Status)

	// Set an extended attribute.
	SetXattr(ino int64, name string, value []byte, flags int) Status

	// Remove an extended attribute.
	RemoveXattr(ino int64, name string) Status
}

type StatVfs struct {
	BlockSize  int64 // Filesystem block size
	Blocks     int64 // Size of filesystem
	BlocksFree int64 // Number of free blocks

	Files     int64 // Number of files
	FilesFree int64 // Number of free inodes

	Fsid    int // Filesystem id
	Flags   FsFlags
	NameMax int // Maximum filename length
}

type DirEntryWriter interface {
	// Returns true if the entry was added, false if there is no more space
	// in the response buffer.
	Add(name string, ino int64, mode int, next int64) bool
}

type FileInfo struct {
	Flags     int
	Writepage bool
	// Bitfields not supported by CGO.
	// TODO: create separate wrapper?
	//DirectIo     bool
	//KeepCache    bool
	//Flush        bool
	//NonSeekable  bool
	//FlockRelease bool
	Handle    uint64
	LockOwner uint64
}

func (f *FileInfo) AccessMode() AccessMode {
	return AccessMode(f.Flags & 3)
}

type ConnInfo struct {
	// Major version of the protocol.
	ProtoMajor int

	// Minor version of the protocol.
	ProtoMinor int

	// Maximum size of the write buffer (writable).
	MaxWrite int

	// Maximum readahead
	MaxReadahead int
}

type Entry struct {
	// Ino is a unique inode number for the filesystem entry.
	//
	// In lookup, zero means negative entry (from version 2.5)
	// Returning ENOENT also means negative entry, but by setting zero
	// ino the kernel may cache negative entries for entry_timeout
	// seconds.
	Ino int64

	// Generation number for this entry.
	//
	// If the file system will be exported over NFS, the
	// ino/generation pairs need to be unique over the file
	// system's lifetime (rather than just the mount time). So if
	// the file system reuses an inode after it has been deleted,
	// it must assign a new, previously unused generation number
	// to the inode at the same time.
	//
	// The generation must be non-zero, otherwise FUSE will treat
	// it as an error.
	Generation int64

	// Inode attributes.
	Attr *InoAttr

	// Validity timeout (in seconds) for the attributes
	AttrTimeout float64

	// Validity timeout (in seconds) for the name
	EntryTimeout float64
}

// Inode attributes.
//
// Even if Timeout == 0, attr must be correct. For example,
// for open(), FUSE uses attr.Size from lookup() to determine
// how many bytes to request. If this value is not correct,
// incorrect data will be returned.
type InoAttr struct {
	Ino   int64
	Size  int64
	Mode  int
	Nlink int

	Uid *int // Defaults to the current uid
	Gid *int // Defaults to the current gid

	Atim time.Time
	Ctim time.Time
	Mtim time.Time

	// Validity timeout (in seconds) for the attributes.
	Timeout float64
}
