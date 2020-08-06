package fuse

// #include "wrapper.h"
// #include <stdlib.h>
import "C"

var fuseSess *C.struct_fuse_session
var fuseChan *C.struct_fuse_chan

// MountAndRun mounts the filesystem and enters the Fuse event loop.
// The argumenst are passed to libfuse to mount the filesystem.  Any flags supported by libfuse are
// allowed. The call returns immediately on error, or else blocks until the filesystem is
// unmounted.
//
// Example:
//
//   fs := &MyFs{}
//   err := fuse.MountAndRun(os.Args, fs)
func MountAndRun(args []string, fs FileSystem) int {
	id := RegisterFS(fs)
	defer DeregisterFS(id)

	// Make args available to C code.
	argv := make([]*C.char, 0, len(args)+1)
	for _, s := range args {
		p := C.CString(s)
		argv = append(argv, p)
	}
	if len(args) < 2 {
		argv = append(argv, C.CString("-h"))
	}
	argc := C.int(len(argv))

	fuseArgs := C.ParseArgs(argc, &argv[0])
	mountpoint := C.ParseMountpoint(fuseArgs)

	ch := C.Mount(mountpoint, fuseArgs)
	if ch == nil {
		return -1
	}

	fuseChan = ch

	se := C.NewSession(C.int(id), fuseArgs, ch)
	if se == nil {
		return -1
	}

	fuseSess = se

	return int(C.Run(se, ch, mountpoint))
}

func UMount(mountpoint string) {
	arg := C.CString(mountpoint)
	C.Exit(fuseSess, fuseChan, arg)
}
