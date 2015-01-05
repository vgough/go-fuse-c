// Provides a CGO wrapper for the FUSE low-level API.
package fuse

// #include "wrapper.h"
// #include <stdlib.h>
import "C"

// TODO: check which operations are provided and tell the C code, so that it only
// sets up bridge methods for implemented operations.
func MountAndRun(args []string, fs RawFileSystem) int {
	id := RegisterRawFs(fs)
	defer DeregisterRawFs(id)

	// Make args available to C code.
	argv := make([]*C.char, len(args)+1)
	for i, s := range args {
		p := C.CString(s)
		argv[i] = p
	}
	argc := C.int(len(args))
	return int(C.MountAndRun(C.int(id), argc, &argv[0], C.getStandardBridgeOps()))
}
