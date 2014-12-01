package main

import (
	"bytes"
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
	now := time.Now()
	m.inodes[1] = &iNode{
		id:    1,
		dir:   root,
		ctime: now,
		mtime: now,
		mode:  0777 | fuse.S_IFDIR,
	}
	return m
}

func (m *MemFs) MakeDir(dir int64, name string, mode int) (*fuse.EntryParam, fuse.Status) {
	n := m.inodes[dir]
	if n == nil {
		return nil, fuse.ENOENT
	}
	d := n.dir
	if d == nil {
		return nil, fuse.ENOTDIR
	}

	if _, exists := d.nodes[name]; exists {
		return nil, fuse.EEXIST
	}

	i := m.nextId
	m.nextId++
	d.nodes[name] = i

	now := time.Now()
	m.inodes[i] = &iNode{
		dir: &memDir{
			parent: n,
			nodes:  make(map[string]int64),
		},
		ctime: now,
		mtime: now,
		mode:  mode | fuse.S_IFDIR,
	}

	e := &fuse.EntryParam{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}
	return e, fuse.OK
}

func (m *MemFs) stat(ino int64) *fuse.InoAttr {
	i := m.inodes[ino]
	if i == nil {
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

	s := m.stat(ino)
	if s == nil {
		return nil, fuse.ENOENT
	} else {
		return s, fuse.OK
	}
}

func (m *MemFs) Lookup(parent int64, name string) (
	entry *fuse.EntryParam, err fuse.Status) {

	n, present := m.inodes[parent]
	if !present {
		return nil, fuse.ENOENT
	}
	if n.dir == nil {
		return nil, fuse.ENOTDIR
	}

	i, present := n.dir.nodes[name]
	if !present {
		return nil, fuse.ENOENT
	}

	e := &fuse.EntryParam{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}

	return e, fuse.OK
}

func (m *MemFs) StatFs(ino int64, s *fuse.StatVfs) fuse.Status {
	s.Files = int64(len(m.inodes))
	return fuse.OK
}

func (m *MemFs) ReadDir(ino int64, fi *fuse.FileInfo, off int64, size int,
	w fuse.DirEntryWriter) fuse.Status {

	n := m.inodes[ino]
	if n == nil {
		return fuse.ENOENT
	}
	d := n.dir
	if d == nil {
		return fuse.ENOTDIR
	}

	idx := int64(1)
	if idx > off {
		if !w.Add(".", ino, n.mode, idx) {
			return fuse.OK
		}
	}
	idx++
	if d.parent != nil {
		if idx > off {
			if !w.Add("..", d.parent.id, d.parent.mode, idx) {
				return fuse.OK
			}
		}
		idx++
	}

	for name, i := range d.nodes {
		if idx > off {
			node := m.inodes[i]
			if !w.Add(name, i, node.mode, idx) {
				return fuse.OK
			}
		}
		idx++
	}
	return fuse.OK
}

func (m *MemFs) Open(ino int64, fi *fuse.FileInfo) fuse.Status {
	n := m.inodes[ino]
	if n == nil {
		return fuse.ENOENT
	}
	if n.dir == nil {
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
	ops := NewMemFs()
	fuse.MountAndRun(args, ops)
}
