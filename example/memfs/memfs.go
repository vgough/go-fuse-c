package main

import (
	"github.com/vgough/go-fuse-c/fuse"
	"os"
)

func main() {
	args := os.Args
	ops := fuse.NewMemFs()
	fuse.MountAndRun(args, ops)
}
