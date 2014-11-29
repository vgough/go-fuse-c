package fuse

import (
	"reflect"
	"time"
	"unsafe"
)

// #include "wrapper.h"
import "C"

func Version() int {
	return int(C.fuse_version())
}

//export ll_Init
func ll_Init(id C.int, cinfo *C.struct_fuse_conn_info) {
	ops := rawFsMap[int(id)]
	info := &ConnInfo{}
	ops.Init(info)
}

//export ll_Destroy
func ll_Destroy(id C.int) {
	ops := rawFsMap[int(id)]
	ops.Destroy()
}

//export ll_Lookup
func ll_Lookup(id C.int, dir C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {

	ops := rawFsMap[int(id)]
	ent, err := ops.Lookup(int64(dir), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_GetAttr
func ll_GetAttr(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info,
	cattr *C.struct_stat, ctimeout *C.double) C.int {

	ops := rawFsMap[int(id)]
	attr, err := ops.GetAttr(int64(ino), newFileInfo(fi))
	if err == OK {
		attr.toCStat(cattr)
		(*ctimeout) = C.double(attr.Timeout)
	}
	return C.int(err)
}

//export ll_ReadDir
func ll_ReadDir(id C.int, ino C.fuse_ino_t, size C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info, db *C.struct_DirBuf) C.int {

	ops := rawFsMap[int(id)]
	writer := &dirBuf{db}
	err := ops.ReadDir(int64(ino), newFileInfo(fi), int64(off), int(size), writer)
	return C.int(err)
}

//export ll_Open
func ll_Open(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	ops := rawFsMap[int(id)]
	info := newFileInfo(fi)
	err := ops.Open(int64(ino), info)
	if err == OK {
		fi.fh = C.uint64_t(info.Handle)
	}
	return C.int(err)
}

//export ll_Read
func ll_Read(id C.int, ino C.fuse_ino_t, off C.off_t,
	fi *C.struct_fuse_file_info, buf unsafe.Pointer, size *C.int) C.int {

	ops := rawFsMap[int(id)]

	// Create slice backed by C buffer.
	hdr := reflect.SliceHeader{
		Data: uintptr(buf),
		Len:  int(*size),
		Cap:  int(*size),
	}
	out := *(*[]byte)(unsafe.Pointer(&hdr))
	n, err := ops.Read(out, int64(ino), int64(off), newFileInfo(fi))
	if err == OK {
		*size = C.int(n)
	}
	return C.int(err)
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
