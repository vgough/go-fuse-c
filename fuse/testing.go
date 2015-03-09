package fuse

import (
	"sync"
	"unsafe"
)

// #include "bridge.h"
// #include <stdlib.h>  // for free()
import "C"

type ReplyHandler func(id int, reply interface{}) int

type ReplyErr struct {
	err Status
}
type ReplyNone struct{}

var reqLoc sync.RWMutex
var reqMap map[int]ReplyHandler = make(map[int]ReplyHandler)
var nextReqId int = 1

func NewReq(handler ReplyHandler, fsId int) C.fuse_req_t {
	reqLoc.Lock()
	defer reqLoc.Unlock()

	id := nextReqId
	nextReqId++
	reqMap[id] = handler
	return C.new_fuse_test_req(C.int(id), C.int(fsId))
}

func FreeReq(r C.fuse_req_t) {
	delete(reqMap, int(C.fuse_test_req_id(r)))
	C.free_fuse_test_req(r)
}

func GetHandler(id C.int) ReplyHandler {
	reqLoc.Lock()
	defer reqLoc.Unlock()

	return reqMap[int(id)]
}

func BridgeLookup(fsId int, ino int64, name string, handler ReplyHandler) {
	req := NewReq(handler, fsId)
	defer FreeReq(req)
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	C.bridge_lookup(req, C.fuse_ino_t(ino), cstr)
}

func BridgeForget(fsId int, ino int64, n int64, handler ReplyHandler) {
	req := NewReq(handler, fsId)
	defer FreeReq(req)
	C.bridge_forget(req, C.fuse_ino_t(ino), C.ulong(n))
}

//export ll_Reply_Err
func ll_Reply_Err(req C.int, err C.int) C.int {
	h := GetHandler(req)
	r := h(int(req), &ReplyErr{Status(err)})
	return C.int(r)
}

//export ll_Reply_None
func ll_Reply_None(req C.int) {
	h := GetHandler(req)
	h(int(req), &ReplyNone{})
}

//export ll_Reply_Entry
func ll_Reply_Entry(req C.int, e *C.struct_fuse_entry_param) C.int {
	return C.int(OK)
}

//export ll_Reply_Create
func ll_Reply_Create(req C.int, e *C.struct_fuse_entry_param,
	fi *C.struct_fuse_file_info) C.int {
	return C.int(OK)
}

//export ll_Reply_Attr
func ll_Reply_Attr(req C.int, attr *C.struct_stat, timeout C.double) C.int {
	return C.int(OK)
}

//export ll_Reply_Readlink
func ll_Reply_Readlink(req C.int, link *C.char) C.int {
	return C.int(OK)
}

//export ll_Reply_Open
func ll_Reply_Open(req C.int, fi *C.struct_fuse_file_info) C.int {
	return C.int(OK)
}

//export ll_Reply_Write
func ll_Reply_Write(req C.int, count C.size_t) C.int {
	return C.int(OK)
}

//export ll_Reply_Buf
func ll_Reply_Buf(req C.int, buf *C.char, size C.size_t) C.int {
	return C.int(OK)
}

//export ll_Reply_Statfs
func ll_Reply_Statfs(req C.int, stbuf *C.struct_statvfs) C.int {
	return C.int(OK)
}

//export ll_Reply_Xattr
func ll_Reply_Xattr(req C.int, size C.size_t) C.int {
	return C.int(OK)
}

//export ll_Add_Direntry
func ll_Add_Direntry(req C.int, buf *C.char, size C.size_t,
	name *C.char, stbuf *C.struct_stat, off C.off_t) C.int {
	return C.int(OK)
}
