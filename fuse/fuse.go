// Provides a CGO wrapper for the FUSE low-level API.
package fuse

// #cgo LDFLAGS: -lfuse
//
// #include "wrapper.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

// Tracks instances of Operations, with a unique identifier used by C code.
// This avoids passing Go pointers into C code.
var opMap map[int]Operations = make(map[int]Operations)
var nextOpId int = 1

// TODO: check which operations are provided and tell the C code, so that it only
// sets up bridge methods for implemented operations.
func MountAndRun(args []string, ops Operations) int {
	id := nextOpId
	nextOpId++
	opMap[id] = ops
	defer delete(opMap, id)

	argv := make([]*C.char, len(args)+1)
	for i, s := range args {
		p := C.CString(s)
		defer C.free(unsafe.Pointer(p))
		argv[i] = p
	}
	argc := C.int(len(args))
	return int(C.MountAndRun(C.int(id), argc, &argv[0]))
}
