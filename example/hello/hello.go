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
	fmt.Println("in Init")
}

func (h *HelloFs) Destroy() {
	fmt.Println("in Destroy")
}

func (h *HelloFs) stat(ino int64) *syscall.Stat_t {
	fmt.Println("stat", ino)
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

	fmt.Println("GetAttr", ino)
	s := h.stat(ino)
	if s == nil {
		return fuse.ENOENT, nil
	} else {
		return fuse.OK, &fuse.Attr{
			Attr:        s,
			AttrTimeout: 1.0,
		}
	}
}

func (h *HelloFs) Lookup(parent int64, name string) (
	err fuse.Status, entry *fuse.EntryParam) {

	fmt.Println("Lookup", parent, name)
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

func (h *HelloFs) ReadDir(ino int64, fi *fuse.FileInfo, off int64, size int,
	w fuse.DirEntryWriter) fuse.Status {

	fmt.Println("ReadDir", ino, off, size)
	if ino != 1 {
		return fuse.ENOTDIR
	}

	if off < 1 {
		w.Add(".", 1, 0, 1)
	}
	if off < 2 {
		w.Add("..", 1, 0, 2)
	}
	if off < 3 {
		w.Add("hello", 2, 0, 3)
	}
	return fuse.OK
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}
