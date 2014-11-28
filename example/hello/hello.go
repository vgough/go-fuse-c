package main

import (
	"fmt"
	"github.com/vgough/go-fuse-c/fuse"
	"os"
	"syscall"
)

const helloStr = "Hello World!\n"

type HelloFs struct {
}

func (h *HelloFs) Init(*fuse.ConnInfo) {
}

func (h *HelloFs) Destroy() {
}

func (h *HelloFs) stat(ino int64) *syscall.Stat_t {
	stat := &syscall.Stat_t{
		Ino: uint64(ino),
	}
	switch ino {
	case 1:
		stat.Mode = fuse.S_IFDIR | 0755
		stat.Nlink = 2
	case 2:
		stat.Mode = fuse.S_IFREG | 0444
		stat.Nlink = 1
		stat.Size = int64(len(helloStr))
	default:
		return nil
	}

	return stat
}

func (h *HelloFs) GetAttr(ino int64, info *fuse.FileInfo) (
	err fuse.Status, attr *fuse.Attr) {
	return fuse.ENOENT, nil
}

func (h *HelloFs) Lookup(parent int64, name string) (
	err fuse.Status, entry *fuse.EntryParam) {

	if parent != 1 || name != "hello" {
		return fuse.ENOENT, nil
	}

	e := &fuse.EntryParam{
		Ino:          2,
		Attr:         h.stat(2),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}

	return fuse.OK, e
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}
