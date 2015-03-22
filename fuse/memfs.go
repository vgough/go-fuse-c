package fuse

import (
	"time"
)

type memDir struct {
	parent *iNode
	nodes  map[string]int64
}

type memFile struct {
	data []byte
}

// iNode is either a directory or a file.
type iNode struct {
	id int64

	// Exactly one of dir,file must be set.
	dir  *memDir
	file *memFile

	ctime time.Time
	mtime time.Time

	// Unix permission bits.
	mode int
}

// MemFs implements an in-memory filesystem.
//
// Directories are stored as an in-memory hash table.  Files are stored as byte arrays.
type MemFs struct {
	DefaultRawFileSystem

	inodes map[int64]*iNode
	nextId int64
}

// NewMemFs creates a new in-memory filesystem.
func NewMemFs() *MemFs {
	root := &memDir{nodes: make(map[string]int64)}
	m := &MemFs{
		inodes: make(map[int64]*iNode),
		nextId: 2, // inode 1 is reserved for the root directory.
	}
	now := time.Now()
	m.inodes[1] = &iNode{
		id:    1,
		dir:   root,
		ctime: now,
		mtime: now,
		mode:  0777 | S_IFDIR,
	}
	return m
}

func (m *MemFs) dirNode(parent int64) (*iNode, Status) {
	n := m.inodes[parent]
	if n == nil {
		return nil, ENOENT
	}
	if n.dir == nil {
		return nil, ENOTDIR
	}
	return n, OK
}

func (m *MemFs) fileNode(ino int64) (*iNode, Status) {
	n := m.inodes[ino]
	if n == nil {
		return nil, ENOENT
	}
	if n.file == nil {
		return nil, EISDIR
	}
	return n, OK
}

func (m *MemFs) Mknod(dir int64, name string, mode int, rdev int) (*Entry, Status) {

	n, err := m.dirNode(dir)
	if err != OK {
		return nil, err
	}

	d := n.dir
	if _, exists := d.nodes[name]; exists {
		return nil, EEXIST
	}

	i := m.nextId
	m.nextId++
	d.nodes[name] = i

	now := time.Now()
	m.inodes[i] = &iNode{
		file: &memFile{
			data: make([]byte, 0),
		},
		ctime: now,
		mtime: now,
		mode:  mode | S_IFREG,
	}

	e := &Entry{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}
	return e, OK
}

func (m *MemFs) Mkdir(dir int64, name string, mode int) (*Entry, Status) {

	n, err := m.dirNode(dir)
	if err != OK {
		return nil, err
	}

	d := n.dir
	if _, exists := d.nodes[name]; exists {
		return nil, EEXIST
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
		mode:  mode | S_IFDIR,
	}

	e := &Entry{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}
	return e, OK
}

func (m *MemFs) stat(ino int64) *InoAttr {
	i := m.inodes[ino]
	if i == nil {
		return nil
	}

	stat := &InoAttr{
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
		stat.Size = int64(len(i.file.data))
	}

	return stat
}

func (m *MemFs) GetAttr(ino int64, info *FileInfo) (attr *InoAttr, err Status) {
	s := m.stat(ino)
	if s == nil {
		return nil, ENOENT
	} else {
		return s, OK
	}
}

func (m *MemFs) SetAttr(ino int64, attr *InoAttr, mask SetAttrMask, fi *FileInfo) (
	*InoAttr, Status) {

	i := m.inodes[ino]
	if i == nil {
		return nil, ENOENT
	}

	if mask&SET_ATTR_MODE != 0 {
		i.mode = attr.Mode
	}
	if mask&SET_ATTR_MTIME != 0 {
		i.mtime = attr.Mtim
	}
	if mask&SET_ATTR_MTIME_NOW != 0 {
		i.mtime = time.Now()
	}

	s := m.stat(ino)
	return s, OK
}

func (m *MemFs) Lookup(parent int64, name string) (entry *Entry, err Status) {
	n, err := m.dirNode(parent)
	if err != OK {
		return nil, err
	}

	i, exist := n.dir.nodes[name]
	if !exist {
		return nil, ENOENT
	}

	e := &Entry{
		Ino:          i,
		Attr:         m.stat(i),
		AttrTimeout:  1.0,
		EntryTimeout: 1.0,
	}

	return e, OK
}

func (m *MemFs) StatFs(ino int64) (stat *StatVfs, status Status) {
	stat = &StatVfs{
		Files: int64(len(m.inodes)),
	}
	status = OK
	return
}

func (m *MemFs) ReadDir(ino int64, fi *FileInfo, off int64, size int, w DirEntryWriter) Status {
	n, err := m.dirNode(ino)
	if err != OK {
		return err
	}
	d := n.dir

	idx := int64(1)
	if idx > off {
		if !w.Add(".", ino, n.mode, idx) {
			return OK
		}
	}
	idx++
	if d.parent != nil {
		if idx > off {
			if !w.Add("..", d.parent.id, d.parent.mode, idx) {
				return OK
			}
		}
		idx++
	}

	for name, i := range d.nodes {
		if idx > off {
			node := m.inodes[i]
			if !w.Add(name, i, node.mode, idx) {
				return OK
			}
		}
		idx++
	}
	return OK
}

func (m *MemFs) Open(ino int64, fi *FileInfo) Status {
	_, err := m.fileNode(ino)
	return err
}

func (m *MemFs) Rmdir(dir int64, name string) Status {
	n, err := m.dirNode(dir)
	if err != OK {
		return err
	}
	cid, present := n.dir.nodes[name]
	if !present {
		return EEXIST
	}

	c := m.inodes[cid]
	if c.dir == nil {
		return ENOTDIR
	}

	if len(c.dir.nodes) > 0 {
		return ENOTEMPTY
	}

	delete(m.inodes, c.id)
	delete(n.dir.nodes, name)
	return OK
}

func (m *MemFs) Rename(dir int64, name string, newdir int64, newname string) Status {
	od, err := m.dirNode(dir)
	if err != OK {
		return err
	}
	oid, present := od.dir.nodes[name]
	if !present {
		return EEXIST
	}

	nd, err := m.dirNode(newdir)
	if err != OK {
		return err
	}

	_, present = nd.dir.nodes[newname]
	nd.dir.nodes[newname] = oid
	delete(od.dir.nodes, name)
	return OK
}

func (m *MemFs) Unlink(dir int64, name string) Status {
	n, err := m.dirNode(dir)
	if err != OK {
		return err
	}
	cid, present := n.dir.nodes[name]
	if !present {
		return EEXIST
	}

	c := m.inodes[cid]
	if c.file == nil {
		return EISDIR
	}

	delete(m.inodes, c.id)
	delete(n.dir.nodes, name)
	return OK
}

func (m *MemFs) Read(p []byte, ino int64, off int64, fi *FileInfo) (int, Status) {
	n, err := m.fileNode(ino)
	if err != OK {
		return 0, err
	}

	data := n.file.data
	l := len(data) - int(off)
	if l >= 0 {
		copy(p, data[off:])
		return l, OK
	} else {
		return 0, OK
	}
}

func (m *MemFs) Write(p []byte, ino int64, off int64, fi *FileInfo) (int, Status) {
	n, err := m.fileNode(ino)
	if err != OK {
		return 0, err
	}

	rl := int(off) + len(p)
	if rl > cap(n.file.data) {
		// Extend
		newSlice := make([]byte, rl)
		copy(newSlice, n.file.data)
		n.file.data = newSlice
	}
	slice := n.file.data[0:rl]
	copy(slice[int(off):rl], p)
	return len(p), OK
}
