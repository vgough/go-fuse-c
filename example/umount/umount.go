package main

import (
	"fmt"
	"os"
	"time"

	"github.com/paddlesteamer/go-fuse-c/fuse"
)

func mount(mountpoint string) {
	args := []string{"", mountpoint}
	ops := &fuse.DefaultFileSystem{}

	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))

	os.Remove(mountpoint)
}

func main() {
	mountpoint := "mnt"

	if err := os.Mkdir(mountpoint, 0755); err != nil {
		panic(err)
	}

	go mount(mountpoint)

	time.Sleep(5 * time.Second)

	fuse.UMount(mountpoint)

	time.Sleep(time.Second)
}
