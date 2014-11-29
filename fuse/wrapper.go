package fuse

import (
	"time"
	"unsafe"
)

// #include "wrapper.h"
import "C"

func Version() int {
	return int(C.fuse_version())
}

//export ll_Init
func ll_Init(t unsafe.Pointer, cinfo *C.struct_fuse_conn_info) {
	ops := (*Operations)(t)
	info := &ConnInfo{}
	(*ops).Init(info)
}

//export ll_Destroy
func ll_Destroy(t unsafe.Pointer) {
	ops := (*Operations)(t)
	(*ops).Destroy()
}

//export ll_Lookup
func ll_Lookup(t unsafe.Pointer, dir C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {

	ops := (*Operations)(t)
	err, ent := (*ops).Lookup(int64(dir), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_GetAttr
func ll_GetAttr(t unsafe.Pointer, ino C.fuse_ino_t, fi *C.struct_fuse_file_info,
	cattr *C.struct_stat, ctimeout *C.double) C.int {

	ops := (*Operations)(t)
	err, attr := (*ops).GetAttr(int64(ino), newFileInfo(fi))
	if err == OK {
		attr.toCStat(cattr)
		(*ctimeout) = C.double(attr.Timeout)
	}
	return C.int(err)
}

const dirBufGrowSize = 8 * 1024

//export ll_ReadDir
func ll_ReadDir(t unsafe.Pointer, ino C.fuse_ino_t, size C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info, db *C.struct_DirBuf) C.int {
	ops := (*Operations)(t)
	writer := &dirBuf{db}
	err := (*ops).ReadDir(int64(ino), newFileInfo(fi), int64(off), int(size), writer)
	return C.int(err)
}

type DirEntryWriter interface {
	Add(name string, ino int64, mode int, next int64) bool
}

type dirBuf struct {
	db *C.struct_DirBuf
}

func (d *dirBuf) Add(name string, ino int64, mode int, next int64) bool {
	// TODO: can we pass pointer to front of name instead of duplicating string?
	cstr := C.CString(name)
	res := C.DirBufAdd(d.db, cstr, C.fuse_ino_t(ino), C.int(mode), C.off_t(next))
	C.free(unsafe.Pointer(cstr))
	return res == 0
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

func newFileInfo(fi *C.struct_fuse_file_info) *FileInfo {
	if fi == nil {
		return nil
	}

	return &FileInfo{
		Flags:     int(fi.flags),
		Writepage: fi.writepage != 0,
		Handle:    uint64(fi.fh),
		LockOwner: uint64(fi.lock_owner),
	}
}

type ConnInfo struct {
	// TODO
}

type EntryParam struct {
	/** Unique inode number
	 *
	 * In lookup, zero means negative entry (from version 2.5)
	 * Returning ENOENT also means negative entry, but by setting zero
	 * ino the kernel may cache negative entries for entry_timeout
	 * seconds.
	 */
	Ino int64

	/** Generation number for this entry.
	 *
	 * If the file system will be exported over NFS, the
	 * ino/generation pairs need to be unique over the file
	 * system's lifetime (rather than just the mount time). So if
	 * the file system reuses an inode after it has been deleted,
	 * it must assign a new, previously unused generation number
	 * to the inode at the same time.
	 *
	 * The generation must be non-zero, otherwise FUSE will treat
	 * it as an error.
	 *
	 */
	Generation int64

	/**
	 * Inode attributes.
	 */
	Attr *InoAttr

	/** Validity timeout (in seconds) for the attributes */
	AttrTimeout float64

	/** Validity timeout (in seconds) for the name */
	EntryTimeout float64
}

/** Inode attributes.
 *
 * Even if Timeout == 0, attr must be correct. For example,
 * for open(), FUSE uses attr.Size from lookup() to determine
 * how many bytes to request. If this value is not correct,
 * incorrect data will be returned.
 */
type InoAttr struct {
	Ino   int64
	Size  int64
	Mode  int
	Nlink int

	Atim time.Time
	Ctim time.Time
	Mtim time.Time

	/** Validity timeout (in seconds) for the attributes */
	Timeout float64
}

func (a *InoAttr) toCStat(o *C.struct_stat) {
	o.st_ino = C.__ino_t(a.Ino)
	o.st_mode = C.__mode_t(a.Mode)
	o.st_nlink = C.__nlink_t(a.Nlink)
	o.st_size = C.__off_t(a.Size)
	toCTime(&o.st_ctim, a.Ctim)
	toCTime(&o.st_mtim, a.Mtim)
	toCTime(&o.st_atim, a.Atim)
}

func toCTime(o *C.struct_timespec, i time.Time) {
	o.tv_sec = C.__time_t(i.Unix())
	o.tv_nsec = C.__syscall_slong_t(i.Nanosecond())
}

func (e *EntryParam) toCEntry(o *C.struct_fuse_entry_param) {
	o.ino = C.fuse_ino_t(e.Ino)
	o.generation = C.ulong(e.Generation)
	e.Attr.toCStat(&o.attr)
	o.attr_timeout = C.double(e.AttrTimeout)
	o.entry_timeout = C.double(e.EntryTimeout)
}

// Operations for Fuse's LowLevel API.
// TODO: allow implementing partial option set.
type Operations interface {
	Init(*ConnInfo)
	Destroy()
	Lookup(dir int64, name string) (err Status, entry *EntryParam)
	GetAttr(ino int64, fi *FileInfo) (err Status, attr *InoAttr)
	ReadDir(ino int64, fi *FileInfo, off int64, size int, w DirEntryWriter) Status
}
