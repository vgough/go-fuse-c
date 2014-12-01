package main

import (
	"bytes"
	"fmt"
	"github.com/vgough/go-fuse-c/fuse"
	"os"
	"time"
)

type memDir struct {
	parent *iNode
	nodes  map[string]int64
}

type memFile struct {
	data bytes.Buffer
}

type iNode struct {
	id int64

	dir  *memDir
	file *memFile

	ctime time.Time
	mtime time.Time

	mode int
}

type MemFs struct {
	fuse.DefaultRawFileSystem

	inodes map[int64]*iNode
	nextId int64
}

func NewMemFs() *MemFs {
	root := &memDir{nodes: make(map[string]int64)}
	m := &MemFs{
		inodes: make(map[int64]*iNode),
		nextId: 2,
	}
	m.inodes[1] = &iNode{
		dir:   root,
		ctime: time.Now(),
		mtime: time.Now(),
		mode:  0777 & fuse.S_IFDIR,
	}
	return m
}

func (m *MemFs) stat(ino int64) *fuse.InoAttr {
	fmt.Println("stat", ino)
	i := m.inodes[ino]
	if i == nil {
		fmt.Println("No such node", ino)
		return nil
	}

	stat := &fuse.InoAttr{
		Ino:     ino,
		Timeout: 1.0,
		Mode:    i.mode,
		Nlink:   1,
		Ctim:    i.ctime,
		Mtim:    i.mtime,
		Atim:    i.mtime,
	}

	if i.dir != nil {
		stat.Size = int64(len(i.dir.nodes))
	} else {
		stat.Size = int64(i.file.data.Len())
	}

	return stat
}

func (m *MemFs) GetAttr(ino int64, info *fuse.FileInfo) (
	attr *fuse.InoAttr, err fuse.Status) {

	fmt.Println("GetAttr", ino)
	s := m.stat(ino)
	if s == nil {
		fmt.Println("ENOENT")
		return nil, fuse.ENOENT
	} else {
		return s, fuse.OK
	}
}

func (m *MemFs) Lookup(parent int64, name string) (
	entry *fuse.EntryParam, err fuse.Status) {

	fmt.Println("Lookup", parent, name)
	n := m.inodes[parent]
	if n.dir == nil {
		fmt.Println("ENOTDIR")
		return nil, fuse.ENOTDIR
	}

	i := n.dir.nodes[name]

	e := &fuse.EntryParam{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}

	return e, fuse.OK
}

func (m *MemFs) StatFs(ino int64, s *fuse.StatVfs) fuse.Status {
	fmt.Println("Statfs", ino)
	s.Files = int64(len(m.inodes))
	return fuse.OK
}

func (m *MemFs) ReadDir(ino int64, fi *fuse.FileInfo, off int64, size int,
	w fuse.DirEntryWriter) fuse.Status {

	fmt.Println("ReadDir", ino)
	n := m.inodes[ino]
	if n == nil {
		fmt.Println("ENOENT")
		return fuse.ENOENT
	}
	d := n.dir
	if d == nil {
		fmt.Println("ENOTDIR")
		return fuse.ENOTDIR
	}

	idx := int64(0)
	if idx >= off {
		w.Add(".", ino, n.mode, 1)
	}
	idx++
	if idx >= off {
		if d.parent != nil {
			w.Add("..", d.parent.id, d.parent.mode, 2)
		}
	}
	idx++

	for name, i := range d.nodes {
		if idx >= off {
			node := m.inodes[i]
			if !w.Add(name, i, node.mode, idx+1) {
				return fuse.OK
			}
		}
		idx++
	}
	return fuse.OK
}

func (m *MemFs) Open(ino int64, fi *fuse.FileInfo) fuse.Status {
	fmt.Println("Open", ino)
	n := m.inodes[ino]
	if n == nil {
		fmt.Println("ENOENT")
		return fuse.ENOENT
	}
	if n.dir == nil {
		fmt.Println("EISDIR")
		return fuse.EISDIR
	}
	return fuse.OK
}

func (m *MemFs) Read(p []byte, ino int64, off int64,
	fi *fuse.FileInfo) (int, fuse.Status) {

	n := m.inodes[ino]
	if n == nil {
		return 0, fuse.ENOENT
	}
	if n.file == nil {
		return 0, fuse.EISDIR
	}

	data := n.file.data.Bytes()
	l := len(data) - int(off)
	if l >= 0 {
		copy(p, data[off:])
		return l, fuse.OK
	} else {
		return 0, fuse.OK
	}
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := NewMemFs()
	fuse.MountAndRun(args, ops)
}
