package main

import (
	"fmt"
	"os"
	"time"

	"github.com/vgough/go-fuse-c/fuse"
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

func (h *HelloFs) StatFs(ino int64) (stat *fuse.StatVfs, err fuse.Status) {
	fmt.Println("statfs", ino)
	stat = &fuse.StatVfs{
		Files:     1,
		FilesFree: 0,
		Flags:     fuse.ST_RDONLY,
	}
	err = fuse.OK
	return
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

func (h *HelloFs) Read(ino int64, size int64, off int64,
	fi *fuse.FileInfo) ([]byte, fuse.Status) {

	fmt.Println("Read", ino, off)
	if ino != 2 {
		return nil, fuse.ENOENT
	}

	data := []byte(helloStr)
	avail := int64(len(data)) - off
	if avail < size {
		size = avail
	}
	if size <= 0 {
		return []byte{}, fuse.OK
	}
	return data[off : off+size], fuse.OK
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}
