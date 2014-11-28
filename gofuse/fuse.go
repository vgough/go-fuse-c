// Gofuse provides a CGO wrapper for the FUSE low-level API.
package gofuse

// #cgo LDFLAGS: -lfuse
//
// #include "wrapper.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

// TODO: check which operations are provided and tell the C code, so that it only
// sets up bridge methods for implemented operations.
func MountAndRun(args []string, ops Operations) int {
	argv := make([]*C.char, len(args)+1)
	for i, s := range args {
		p := C.CString(s)
		defer C.free(unsafe.Pointer(p))
		argv[i] = p
	}
	argc := C.int(len(args))
	return int(C.MountAndRun(unsafe.Pointer(&ops), argc, &argv[0]))
}
