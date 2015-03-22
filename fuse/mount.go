package fuse

// #include "wrapper.h"
// #include <stdlib.h>
import "C"

// MountAndRun mounts the filesystem and enters the Fuse event loop.
// The argumenst are passed to libfuse to mount the filesystem.  Any flags supported by libfuse are
// allowed. The call returns immediately on error, or else blocks until the filesystem is
// unmounted.
//
// Example:
//
//   fs := &MyFs{}
//   err := fuse.MountAndRun(os.Args, fs)
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
