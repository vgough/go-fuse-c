package fuse

// #include "wrapper.h"
// #include <stdlib.h>
import "C"
import (
	"os"
	"path/filepath"
)

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

	fuseArgs := C.InitArgs(argc, &argv[0])
	mountpoint := C.ParseMountpoint(fuseArgs)

	ch := C.Mount(mountpoint, fuseArgs)
	if ch == nil {
		return -1
	}

	se := C.NewSession(mountpoint, fuseArgs, ch)
	if se == nil {
		return -1
	}

	mp := C.GoString(mountpoint)

	RegisterFS(mp, fs, se, ch)
	defer DeregisterFS(mp)

	return int(C.Run(mountpoint, se, ch))
}

func UMount(mountpoint string) {
	if !filepath.IsAbs(mountpoint) {
		cwd, _ := os.Getwd()

		mountpoint = filepath.Join(cwd, mountpoint)
	}

	mountpoint, _ = filepath.Abs(mountpoint)

	minfo := getMountInfo(mountpoint)

	C.Exit(C.CString(mountpoint), minfo.se, minfo.ch)
}
