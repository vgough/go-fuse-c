package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/paddlesteamer/go-fuse-c/fuse"
)

func mount(mountpoint string, wg *sync.WaitGroup) {
	defer wg.Done()

	args := []string{"", mountpoint}
	ops := &fuse.DefaultFileSystem{}

	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}

func main() {
	mountpoint := "mnt"

	if err := os.Mkdir(mountpoint, 0755); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go mount(mountpoint, &wg)

	time.Sleep(3 * time.Second)

	fmt.Println("umount")
	fuse.UMount(mountpoint)
	fmt.Println("umount done")

	wg.Wait()

	os.Remove(mountpoint)
}
