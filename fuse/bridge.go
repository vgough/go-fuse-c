package fuse

import (
	"reflect"
	"sync"
	"time"
	"unsafe"
)

// #include "wrapper.h"
// #include <stdlib.h>  // for free()
import "C"

// State which tracks instances of FileSystem, with a unique identifier used
// by C code.  This avoids passing Go pointers into C code.
var fsMapLock sync.RWMutex
var rawFSMap = make(map[string]MountInfo)

type MountInfo struct {
	fs FileSystem
	se *C.struct_fuse_session
	ch *C.struct_fuse_chan
}

// enableBridgeTestMode can be used to enable the global bridge test mode.
// This prevents fuse_reply callbacks, since there is no active FUSE filesystem.
// Used to allow internal testing of C <-> Go translation layers.
func enableBridgeTestMode() {
	C.enable_bridge_test_mode()
}

// RegisterFS registers a filesystem with the bridge layer.
// Returns an integer id, which identifies the filesystem instance.
//
// When calling the FUSE lowlevel initialization method (eg fuse_lowlevel_new), the userdata
// argument must be a pointer to an integer holding this id value.  The bridge methods use this to
// determine which filesystem will handle FUSE callbacks.
//
// When the filesystem is no longer active, DeregisterFS can be called to release resources.
func RegisterFS(mountpoint string, fs FileSystem, se *C.struct_fuse_session, ch *C.struct_fuse_chan) {
	fsMapLock.Lock()
	defer fsMapLock.Unlock()

	rawFSMap[mountpoint] = MountInfo{
		fs: fs,
		se: se,
		ch: ch,
	}
}

// DeregisterFS releases a previously allocated filesystem from RegisterRawFs.
func DeregisterFS(mountpoint string) {
	fsMapLock.Lock()
	defer fsMapLock.Unlock()

	delete(rawFSMap, mountpoint)
}

// getFS returns the filesystem for the given mountpoint.
func getFS(mountpoint string) FileSystem {
	fsMapLock.RLock()
	defer fsMapLock.RUnlock()

	mi := rawFSMap[mountpoint]
	return mi.fs
}

func getMountInfo(mountpoint string) MountInfo {
	fsMapLock.RLock()
	defer fsMapLock.RUnlock()

	return rawFSMap[mountpoint]
}

// Version returns the version number from the linked libfuse client implementation.
func Version() int {
	return int(C.fuse_version())
}

// zeroCopyBuf creates a byte array backed by a C buffer.
func zeroCopyBuf(buf unsafe.Pointer, size int) []byte {
	// Create slice backed by C buffer.
	hdr := reflect.SliceHeader{
		Data: uintptr(buf),
		Len:  size,
		Cap:  size,
	}
	return *(*[]byte)(unsafe.Pointer(&hdr))
}

//export ll_Init
func ll_Init(mountpoint *C.char, cinfo *C.struct_fuse_conn_info) {
	fs := getFS(C.GoString(mountpoint))
	info := &ConnInfo{
		ProtoMajor:   int(cinfo.proto_major),
		ProtoMinor:   int(cinfo.proto_minor),
		MaxWrite:     int(cinfo.max_write),
		MaxReadahead: int(cinfo.max_readahead),
	}
	fs.Init(info)

	// Copy writable options back to cinfo
	cinfo.max_write = C.uint(info.MaxWrite)
	cinfo.max_readahead = C.uint(info.MaxReadahead)

	// TODO: async_read
	// TODO: APPLE specific flag support.
}

//export ll_Destroy
func ll_Destroy(mountpoint *C.char) {
	fs := getFS(C.GoString(mountpoint))
	fs.Destroy()
}

//export ll_StatFS
func ll_StatFS(mountpoint *C.char, ino C.fuse_ino_t, stat *C.struct_statvfs) C.int {
	fs := getFS(C.GoString(mountpoint))
	s, err := fs.StatFS(int64(ino))
	if err == OK {
		s.toCStat(stat)
	}
	return C.int(err)
}

//export ll_SetXAttr
func ll_SetXAttr(mountpoint *C.char, ino C.fuse_ino_t, name *C.char, value unsafe.Pointer,
	size C.size_t, flags C.int) C.int {

	fs := getFS(C.GoString(mountpoint))
	data := zeroCopyBuf(value, int(size))
	err := fs.SetXAttr(int64(ino), C.GoString(name), data, int(flags))
	return C.int(err)
}

//export ll_GetXAttr
func ll_GetXAttr(mountpoint *C.char, ino C.fuse_ino_t, name *C.char, buf unsafe.Pointer,
	size *C.size_t) C.int {

	fs := getFS(C.GoString(mountpoint))
	var err Status
	var outSize int
	if *size == 0 {
		outSize, err = fs.GetXAttrSize(int64(ino), C.GoString(name))
	} else {
		out := zeroCopyBuf(buf, int(*size))
		outSize, err = fs.GetXAttr(int64(ino), C.GoString(name), out)
	}
	if err == OK {
		*size = C.size_t(outSize)
	}
	return C.int(err)
}

//export ll_Lookup
func ll_Lookup(mountpoint *C.char, dir C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {

	fs := getFS(C.GoString(mountpoint))
	ent, err := fs.Lookup(int64(dir), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Forget
func ll_Forget(mountpoint *C.char, ino C.fuse_ino_t, n C.int) {
	fs := getFS(C.GoString(mountpoint))
	fs.Forget(int64(ino), int(n))
}

//export ll_GetAttr
func ll_GetAttr(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info,
	cattr *C.struct_stat, ctimeout *C.double) C.int {

	fs := getFS(C.GoString(mountpoint))
	attr, err := fs.GetAttr(int64(ino), newFileInfo(fi))
	if err == OK {
		attr.toCStat(cattr, ctimeout)
	}
	return C.int(err)
}

//export ll_SetAttr
func ll_SetAttr(mountpoint *C.char, ino C.fuse_ino_t, attr *C.struct_stat, toSet C.int,
	fi *C.struct_fuse_file_info, cattr *C.struct_stat, ctimeout *C.double) C.int {

	fs := getFS(C.GoString(mountpoint))
	var ia InoAttr
	ia.fromCStat(attr)
	oattr, err := fs.SetAttr(int64(ino), &ia, SetAttrMask(toSet), newFileInfo(fi))
	if err == OK {
		oattr.toCStat(cattr, ctimeout)
	}
	return C.int(err)
}

//export ll_ReadDir
func ll_ReadDir(mountpoint *C.char, ino C.fuse_ino_t, size C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info, db *C.struct_DirBuf) C.int {

	fs := getFS(C.GoString(mountpoint))
	writer := &dirBuf{db}
	err := fs.ReadDir(int64(ino), newFileInfo(fi), int64(off), int(size), writer)
	return C.int(err)
}

//export ll_Open
func ll_Open(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	info := newFileInfo(fi)
	err := fs.Open(int64(ino), info)
	if err == OK {
		fi.fh = C.uint64_t(info.Handle)
	}
	return C.int(err)
}

//export ll_OpenDir
func ll_OpenDir(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	info := newFileInfo(fi)
	err := fs.OpenDir(int64(ino), info)
	if err == OK {
		fi.fh = C.uint64_t(info.Handle)
	}
	return C.int(err)
}

//export ll_Release
func ll_Release(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.Release(int64(ino), newFileInfo(fi))
	return C.int(err)
}

//export ll_ReleaseDir
func ll_ReleaseDir(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.ReleaseDir(int64(ino), newFileInfo(fi))
	return C.int(err)
}

//export ll_FSync
func ll_FSync(mountpoint *C.char, ino C.fuse_ino_t, datasync C.int, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	var dataOnly bool = datasync != 0
	err := fs.FSync(int64(ino), dataOnly, newFileInfo(fi))
	return C.int(err)
}

//export ll_FSyncDir
func ll_FSyncDir(mountpoint *C.char, ino C.fuse_ino_t, datasync C.int, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	var dataOnly bool = datasync != 0
	err := fs.FSyncDir(int64(ino), dataOnly, newFileInfo(fi))
	return C.int(err)
}

//export ll_Flush
func ll_Flush(mountpoint *C.char, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.Flush(int64(ino), newFileInfo(fi))
	return C.int(err)
}

//export ll_Read
func ll_Read(mountpoint *C.char, req C.fuse_req_t, ino C.fuse_ino_t, size C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info) C.int {

	fs := getFS(C.GoString(mountpoint))

	// Create slice backed by C buffer.
	buf, err := fs.Read(int64(ino), int64(size), int64(off), newFileInfo(fi))
	if err != OK {
		return C.int(err)
	}

	ptr := unsafe.Pointer(&buf[0])
	return C.reply_buf(req, (*C.char)(ptr), size)
}

//export ll_Write
func ll_Write(mountpoint *C.char, ino C.fuse_ino_t, buf unsafe.Pointer, n *C.size_t, off C.off_t,
	fi *C.struct_fuse_file_info) C.int {

	fs := getFS(C.GoString(mountpoint))
	in := zeroCopyBuf(buf, int(*n))
	written, err := fs.Write(in, int64(ino), int64(off), newFileInfo(fi))
	if err == OK {
		*n = C.size_t(written)
	}
	return C.int(err)
}

//export ll_Mknod
func ll_Mknod(mountpoint *C.char, dir C.fuse_ino_t, name *C.char, mode C.mode_t,
	rdev C.dev_t, cent *C.struct_fuse_entry_param) C.int {

	fs := getFS(C.GoString(mountpoint))
	ent, err := fs.Mknod(int64(dir), C.GoString(name), int(mode), int(rdev))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_ListXAttr
func ll_ListXAttr(mountpoint *C.char, ino C.fuse_ino_t, buf unsafe.Pointer, size *C.size_t) C.int {
	out := zeroCopyBuf(buf, int(*size))
	fs := getFS(C.GoString(mountpoint))
	keys, err := fs.ListXAttrs(int64(ino))
	if err != OK {
		return C.int(err)
	}

	var pos int
	for _, k := range keys {
		totalLen := pos + len(k) + 1
		if totalLen < int(*size) {
			copy(out[pos:totalLen], k)
			out[totalLen] = 0
			pos = totalLen
		} else {
			*size = C.size_t(pos)
			return C.int(ERANGE)
		}
	}

	*size = C.size_t(pos)
	return C.int(OK)
}

//export ll_RemoveXAttr
func ll_RemoveXAttr(mountpoint *C.char, ino C.fuse_ino_t, name *C.char) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.RemoveXAttr(int64(ino), C.GoString(name))
	return C.int(err)
}

//export ll_Access
func ll_Access(mountpoint *C.char, ino C.fuse_ino_t, mask C.int) C.int {
	fs := getFS(C.GoString(mountpoint))
	return C.int(fs.Access(int64(ino), int(mask)))
}

//export ll_Create
func ll_Create(mountpoint *C.char, dir C.fuse_ino_t, name *C.char, mode C.mode_t,
	fi *C.struct_fuse_file_info, cent *C.struct_fuse_entry_param) C.int {

	fs := getFS(C.GoString(mountpoint))
	info := newFileInfo(fi)
	ent, err := fs.Create(int64(dir), C.GoString(name), int(mode), info)
	if err == OK {
		ent.toCEntry(cent)
		fi.fh = C.uint64_t(info.Handle)
	}
	return C.int(err)
}

//export ll_Mkdir
func ll_Mkdir(mountpoint *C.char, dir C.fuse_ino_t, name *C.char, mode C.mode_t,
	cent *C.struct_fuse_entry_param) C.int {

	fs := getFS(C.GoString(mountpoint))
	ent, err := fs.Mkdir(int64(dir), C.GoString(name), int(mode))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Rmdir
func ll_Rmdir(mountpoint *C.char, dir C.fuse_ino_t, name *C.char) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.Rmdir(int64(dir), C.GoString(name))
	return C.int(err)
}

//export ll_Symlink
func ll_Symlink(mountpoint *C.char, link *C.char, parent C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {
	fs := getFS(C.GoString(mountpoint))
	ent, err := fs.Symlink(C.GoString(link), int64(parent), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_Link
func ll_Link(mountpoint *C.char, ino C.fuse_ino_t, newparent C.fuse_ino_t, name *C.char,
	cent *C.struct_fuse_entry_param) C.int {
	fs := getFS(C.GoString(mountpoint))
	ent, err := fs.Link(int64(ino), int64(newparent), C.GoString(name))
	if err == OK {
		ent.toCEntry(cent)
	}
	return C.int(err)
}

//export ll_ReadLink
func ll_ReadLink(mountpoint *C.char, ino C.fuse_ino_t, err *C.int) *C.char {
	fs := getFS(C.GoString(mountpoint))
	s, e := fs.ReadLink(int64(ino))
	*err = C.int(e)
	if e == OK {
		return C.CString(s)
	}
	return nil
}

//export ll_Unlink
func ll_Unlink(mountpoint *C.char, dir C.fuse_ino_t, name *C.char) C.int {
	fs := getFS(C.GoString(mountpoint))
	err := fs.Unlink(int64(dir), C.GoString(name))
	return C.int(err)
}

//export ll_Rename
func ll_Rename(mountpoint *C.char, dir C.fuse_ino_t, name *C.char,
	newdir C.fuse_ino_t, newname *C.char) C.int {

	fs := getFS(C.GoString(mountpoint))
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

func (e *Entry) toCEntry(o *C.struct_fuse_entry_param) {
	o.ino = C.fuse_ino_t(e.Ino)
	o.generation = C.ulong(e.Generation)
	e.Attr.toCStat(&o.attr, nil)
	o.attr_timeout = C.double(e.AttrTimeout)
	o.entry_timeout = C.double(e.EntryTimeout)
}

// Use C wrapper function to avoid issues with different typedef names on different systems.
func toCTime(o *C.struct_timespec, i time.Time) {
	C.FillTimespec(o, C.time_t(i.Unix()), C.ulong(i.Nanosecond()))
}
