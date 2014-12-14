package main

import (
	"fmt"
	"github.com/vgough/go-fuse-c/fuse"
	"os"
	"time"
)

const helloStr = "Hello World!\n"

var mountTime = time.Now()

type HelloFs struct {
	fuse.DefaultRawFileSystem
}

func (h *HelloFs) stat(ino int64) *fuse.InoAttr {
	fmt.Println("stat", ino)
	stat := &fuse.InoAttr{
		Ino:     ino,
		Timeout: 1.0,
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

	stat.Atim = mountTime
	stat.Ctim = mountTime
	stat.Mtim = mountTime

	return stat
}

func (h *HelloFs) GetAttr(ino int64, info *fuse.FileInfo) (
	attr *fuse.InoAttr, err fuse.Status) {

	fmt.Println("GetAttr", ino)
	s := h.stat(ino)
	if s == nil {
		return nil, fuse.ENOENT
	} else {
		return s, fuse.OK
	}
}

func (h *HelloFs) Lookup(parent int64, name string) (
	entry *fuse.Entry, err fuse.Status) {

	fmt.Println("Lookup", parent, name)
	if parent != 1 || name != "hello" {
		return nil, fuse.ENOENT
	}

	e := &fuse.Entry{
		Ino:          2,
		Attr:         h.stat(2),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}

	return e, fuse.OK
}

func (h *HelloFs) StatFs(ino int64, s *fuse.StatVfs) fuse.Status {
	fmt.Println("statfs", ino)
	s.Files = 1
	s.FilesFree = 0
	s.Flags = fuse.ST_RDONLY
	return fuse.OK
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

func (h *HelloFs) Open(ino int64, fi *fuse.FileInfo) fuse.Status {
	fmt.Println("Open", ino)
	if ino != 2 {
		return fuse.EISDIR
	} else if fi.AccessMode() != fuse.O_RDONLY {
		return fuse.EACCES
	} else {
		return fuse.OK
	}
}

func (h *HelloFs) Read(p []byte, ino int64, off int64,
	fi *fuse.FileInfo) (int, fuse.Status) {

	fmt.Println("Read", ino, off)
	if ino != 2 {
		return 0, fuse.ENOENT
	}

	data := []byte(helloStr)
	n := len(data) - int(off)
	if n < 0 {
		n = 0
	}
	copy(p, data[off:])
	return n, fuse.OK
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}
