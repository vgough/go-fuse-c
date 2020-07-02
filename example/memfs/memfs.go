package main

import (
	"log"
	"os"

	"github.com/paddlesteamer/go-fuse-c/fuse"
)

func main() {
	args := os.Args
	ops := fuse.NewMemFs()
	if res := fuse.MountAndRun(args, ops); res < 0 {
		log.Fatalf("failed with error: %d", res)
	}
}
