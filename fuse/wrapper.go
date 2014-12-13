package fuse

import (
	"reflect"
	"time"
	"unsafe"
)

// #include "wrapper.h"
// #include <stdlib.h>  // for free()
import "C"

func Version() int {
	return int(C.fuse_version())
}

//export ll_Init
func ll_Init(id C.int, cinfo *C.struct_fuse_conn_info) {
	fs := rawFsMap[int(id)]
	info := &ConnInfo{}
	fs.Init(info)
}

//export ll_Destroy
func ll_Destroy(id C.int) {
	fs := rawFsMap[int(id)]
	fs.Destroy()
}

//export ll_StatFs
func ll_StatFs(id C.int, ino C.fuse_ino_t, stat *C.struct_statvfs) C.int {
	fs := rawFsMap[int(id)]
	var s StatVfs
	err := fs.StatFs(int64(ino), &s)
	if err == OK {
		s.toCStat(stat)
	}
	return C.int(err)
}

//export ll_Lookup
func ll_Lookup(id C.int, dir C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {

	fs := rawFsMap[int(id)]
	ent, err := fs.Lookup(int64(dir), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Forget
func ll_Forget(id C.int, ino C.fuse_ino_t, n C.int) {
	fs := rawFsMap[int(id)]
	fs.Forget(int64(ino), int(n))
}

//export ll_GetAttr
func ll_GetAttr(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info,
	cattr *C.struct_stat, ctimeout *C.double) C.int {

	fs := rawFsMap[int(id)]
	attr, err := fs.GetAttr(int64(ino), newFileInfo(fi))
	if err == OK {
		attr.toCStat(cattr, ctimeout)
	}
	return C.int(err)
}

//export ll_SetAttr
func ll_SetAttr(id C.int, ino C.fuse_ino_t, attr *C.struct_stat, toSet C.int,
	fi *C.struct_fuse_file_info, cattr *C.struct_stat, ctimeout *C.double) C.int {

	fs := rawFsMap[int(id)]
	var ia InoAttr
	ia.fromCStat(attr)
	oattr, err := fs.SetAttr(int64(ino), &ia, SetAttrMask(toSet), newFileInfo(fi))
	if err == OK {
		oattr.toCStat(cattr, ctimeout)
	}
	return C.int(err)
}

//export ll_ReadDir
func ll_ReadDir(id C.int, ino C.fuse_ino_t, size C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info, db *C.struct_DirBuf) C.int {

	fs := rawFsMap[int(id)]
	writer := &dirBuf{db}
	err := fs.ReadDir(int64(ino), newFileInfo(fi), int64(off), int(size), writer)
	return C.int(err)
}

//export ll_Open
func ll_Open(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := rawFsMap[int(id)]
	info := newFileInfo(fi)
	err := fs.Open(int64(ino), info)
	if err == OK {
		fi.fh = C.uint64_t(info.Handle)
	}
	return C.int(err)
}

//export ll_Read
func ll_Read(id C.int, ino C.fuse_ino_t, off C.off_t,
	fi *C.struct_fuse_file_info, buf unsafe.Pointer, size *C.int) C.int {

	fs := rawFsMap[int(id)]

	// Create slice backed by C buffer.
	hdr := reflect.SliceHeader{
		Data: uintptr(buf),
		Len:  int(*size),
		Cap:  int(*size),
	}
	out := *(*[]byte)(unsafe.Pointer(&hdr))
	n, err := fs.Read(out, int64(ino), int64(off), newFileInfo(fi))
	if err == OK {
		*size = C.int(n)
	}
	return C.int(err)
}

//export ll_Write
func ll_Write(id C.int, ino C.fuse_ino_t, buf unsafe.Pointer, n *C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info) C.int {

	fs := rawFsMap[int(id)]
	// Create slice backed by C buffer.
	hdr := reflect.SliceHeader{
		Data: uintptr(buf),
		Len:  int(*n),
		Cap:  int(*n),
	}
	in := *(*[]byte)(unsafe.Pointer(&hdr))
	written, err := fs.Write(in, int64(ino), int64(off), newFileInfo(fi))
	if err == OK {
		*n = C.size_t(written)
	}
	return C.int(err)
}

//export ll_Mknod
func ll_Mknod(id C.int, dir C.fuse_ino_t, name *C.char, mode C.mode_t,
	rdev C.dev_t, cent *C.struct_fuse_entry_param) C.int {

	fs := rawFsMap[int(id)]
	ent, err := fs.Mknod(int64(dir), C.GoString(name), int(mode), int(rdev))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Mkdir
func ll_Mkdir(id C.int, dir C.fuse_ino_t, name *C.char, mode C.mode_t,
	cent *C.struct_fuse_entry_param) C.int {

	fs := rawFsMap[int(id)]
	ent, err := fs.Mkdir(int64(dir), C.GoString(name), int(mode))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Rmdir
func ll_Rmdir(id C.int, dir C.fuse_ino_t, name *C.char) C.int {
	fs := rawFsMap[int(id)]
	err := fs.Rmdir(int64(dir), C.GoString(name))
	return C.int(err)
}

//export ll_Unlink
func ll_Unlink(id C.int, dir C.fuse_ino_t, name *C.char) C.int {
	fs := rawFsMap[int(id)]
	err := fs.Unlink(int64(dir), C.GoString(name))
	return C.int(err)
}

//export ll_Rename
func ll_Rename(id C.int, dir C.fuse_ino_t, name *C.char,
	newdir C.fuse_ino_t, newname *C.char) C.int {

	fs := rawFsMap[int(id)]
	err := fs.Rename(int64(dir), C.GoString(name), int64(newdir), C.GoString(newname))
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

func (s *StatVfs) toCStat(o *C.struct_statvfs) {
	o.f_bsize = C.ulong(s.BlockSize)
	o.f_blocks = C.__fsblkcnt64_t(s.Blocks)
	o.f_bfree = C.__fsblkcnt64_t(s.BlocksFree)

	o.f_files = C.__fsfilcnt64_t(s.Files)
	o.f_ffree = C.__fsfilcnt64_t(s.FilesFree)

	o.f_fsid = C.ulong(s.Fsid)
	o.f_flag = C.ulong(s.Flags)
	o.f_namemax = C.ulong(s.NameMax)
}

func (a *InoAttr) fromCStat(i *C.struct_stat) {
	a.Ino = int64(i.st_ino)
	a.Mode = int(i.st_mode)
	a.Nlink = int(i.st_nlink)
	a.Size = int64(i.st_size)
	var uid int = int(i.st_uid)
	var gid int = int(i.st_gid)
	a.Uid = &uid
	a.Gid = &gid
	a.Atim = time.Unix(int64(i.st_atim.tv_sec), int64(i.st_atim.tv_nsec))
	a.Ctim = time.Unix(int64(i.st_ctim.tv_sec), int64(i.st_ctim.tv_nsec))
	a.Mtim = time.Unix(int64(i.st_mtim.tv_sec), int64(i.st_mtim.tv_nsec))
}

func (a *InoAttr) toCStat(o *C.struct_stat, timeout *C.double) {
	o.st_ino = C.__ino_t(a.Ino)
	o.st_mode = C.__mode_t(a.Mode)
	o.st_nlink = C.__nlink_t(a.Nlink)
	o.st_size = C.__off_t(a.Size)
	if a.Uid != nil {
		o.st_uid = C.__uid_t(*a.Uid)
	}
	if a.Gid != nil {
		o.st_gid = C.__gid_t(*a.Gid)
	}
	toCTime(&o.st_ctim, a.Ctim)
	toCTime(&o.st_mtim, a.Mtim)
	toCTime(&o.st_atim, a.Atim)
	if timeout != nil {
		(*timeout) = C.double(a.Timeout)
	}
}

func toCTime(o *C.struct_timespec, i time.Time) {
	o.tv_sec = C.__time_t(i.Unix())
	o.tv_nsec = C.__syscall_slong_t(i.Nanosecond())
}

func (e *EntryParam) toCEntry(o *C.struct_fuse_entry_param) {
	o.ino = C.fuse_ino_t(e.Ino)
	o.generation = C.ulong(e.Generation)
	e.Attr.toCStat(&o.attr, nil)
	o.attr_timeout = C.double(e.AttrTimeout)
	o.entry_timeout = C.double(e.EntryTimeout)
}
