package main

import (
	"os"

	"github.com/vgough/go-fuse-c/fuse"
)

func main() {
	args := os.Args
	ops := fuse.NewMemFs()
	fuse.MountAndRun(args, ops)
}
