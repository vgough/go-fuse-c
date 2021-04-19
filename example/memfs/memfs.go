package main

import (
	"log"
	"os"

	"github.com/vgough/go-fuse-c/fuse"
)

func main() {
	args := os.Args
	// The implementation lives in fuse/memfs, because an in-memory filesystem
	// is also useful in testing.
	ops := fuse.NewMemFS()
	if res := fuse.MountAndRun(args, ops); res < 0 {
		log.Fatalf("failed with error: %d", res)
	}
}
