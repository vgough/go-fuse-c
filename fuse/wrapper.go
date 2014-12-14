package fuse

import (
	"reflect"
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

//export ll_Release
func ll_Release(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := rawFsMap[int(id)]
	err := fs.Release(int64(ino), newFileInfo(fi))
	return C.int(err)
}

//export ll_FSync
func ll_FSync(id C.int, ino C.fuse_ino_t, datasync C.int, fi *C.struct_fuse_file_info) C.int {
	fs := rawFsMap[int(id)]
	err := fs.FSync(int64(ino), int(datasync), newFileInfo(fi))
	return C.int(err)
}

//export ll_Flush
func ll_Flush(id C.int, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := rawFsMap[int(id)]
	err := fs.Flush(int64(ino), newFileInfo(fi))
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

//export ll_Symlink
func ll_Symlink(id C.int, link *C.char, parent C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {
	fs := rawFsMap[int(id)]
	ent, err := fs.Symlink(C.GoString(link), int64(parent), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Link
func ll_Link(id C.int, ino C.fuse_ino_t, newparent C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {
	fs := rawFsMap[int(id)]
	ent, err := fs.Link(int64(ino), int64(newparent), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_ReadLink
func ll_ReadLink(id C.int, ino C.fuse_ino_t, err *C.int) *C.char {
	fs := rawFsMap[int(id)]
	s, e := fs.ReadLink(int64(ino))
	*err = C.int(e)
	if e == OK {
		return C.CString(s)
	} else {
		return nil
	}
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

func (e *EntryParam) toCEntry(o *C.struct_fuse_entry_param) {
	o.ino = C.fuse_ino_t(e.Ino)
	o.generation = C.ulong(e.Generation)
	e.Attr.toCStat(&o.attr, nil)
	o.attr_timeout = C.double(e.AttrTimeout)
	o.entry_timeout = C.double(e.EntryTimeout)
}
