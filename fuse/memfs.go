package fuse

import (
	"log/slog"
	"time"
)

const attrTimeout = 60.0

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

func (i *iNode) stat() *InoAttr {
	stat := &InoAttr{
		Ino:     i.id,
		Timeout: attrTimeout,
		Mode:    i.mode,
		NLink:   1,
		CTime:   i.ctime,
		MTime:   i.mtime,
		ATime:   i.mtime,
	}

	if i.dir != nil {
		stat.Size = int64(len(i.dir.nodes))
		stat.Mode |= S_IFDIR
	} else {
		stat.Size = int64(len(i.file.data))
		stat.Mode |= S_IFREG
	}
	return stat
}

// MemFS implements an in-memory filesystem.
//
// Directories are stored as an in-memory hash table.  Files are stored as byte arrays.
type MemFS struct {
	DefaultFileSystem

	gen    uint64
	inodes map[int64]*iNode
	nextID int64
}

// NewMemFS creates a new in-memory filesystem.
func NewMemFS() *MemFS {
	root := &memDir{nodes: make(map[string]int64)}
	m := &MemFS{
		inodes: make(map[int64]*iNode),
		nextID: 2, // inode 1 is reserved for the root directory.
		gen:    uint64(time.Now().UnixMicro()),
	}
	now := time.Now()
	m.inodes[1] = &iNode{
		id:    1,
		dir:   root,
		ctime: now,
		mtime: now,
		mode:  0o777,
	}
	return m
}

func (m *MemFS) entry(node *iNode) *Entry {
	return &Entry{
		Ino:          node.id,
		Generation:   m.gen,
		Attr:         node.stat(),
		AttrTimeout:  attrTimeout,
		EntryTimeout: attrTimeout,
	}
}

func (m *MemFS) dirNode(parent int64) (*iNode, Status) {
	n := m.inodes[parent]
	if n == nil {
		return nil, ENOENT
	}
	if n.dir == nil {
		return nil, ENOTDIR
	}
	return n, OK
}

func (m *MemFS) fileNode(ino int64) (*iNode, Status) {
	n := m.inodes[ino]
	if n == nil {
		return nil, ENOENT
	}
	if n.file == nil {
		return nil, EISDIR
	}
	return n, OK
}

// Mknod creates nodes.
func (m *MemFS) Mknod(dir int64, name string, mode int, rdev int) (*Entry, Status) {
	slog.Debug("Mknod", "dir", dir, "name", name, "mode", mode, "rdev", rdev)

	n, err := m.dirNode(dir)
	if err != OK {
		return nil, err
	}

	d := n.dir
	if _, exists := d.nodes[name]; exists {
		return nil, EEXIST
	}

	i := m.nextID
	m.nextID++
	d.nodes[name] = i

	now := time.Now()
	node := &iNode{
		id: i,
		file: &memFile{
			data: make([]byte, 0),
		},
		ctime: now,
		mtime: now,
		mode:  mode,
	}
	m.inodes[i] = node
	return m.entry(node), OK
}

func (m *MemFS) Create(dir int64, name string, mode int, fi *FileInfo) (*Entry, Status) {
	return m.Mknod(dir, name, mode, 0)
}

// Mkdir create directories.
func (m *MemFS) Mkdir(dir int64, name string, mode int) (*Entry, Status) {
	slog.Debug("Mkdir", "dir", dir, "name", name, "mode", mode)

	n, err := m.dirNode(dir)
	if err != OK {
		return nil, err
	}

	d := n.dir
	if _, exists := d.nodes[name]; exists {
		return nil, EEXIST
	}

	i := m.nextID
	m.nextID++
	d.nodes[name] = i

	now := time.Now()
	node := &iNode{
		id: i,
		dir: &memDir{
			parent: n,
			nodes:  make(map[string]int64),
		},
		ctime: now,
		mtime: now,
		mode:  mode,
	}
	m.inodes[i] = node
	return m.entry(node), OK
}

// GetAttr returns node attributes.
func (m *MemFS) GetAttr(ino int64, info *FileInfo) (attr *InoAttr, err Status) {
	slog.Debug("GetAttr", "ino", ino)

	i := m.inodes[ino]
	if i == nil {
		return nil, ENOENT
	}
	return i.stat(), OK
}

// SetAttr changes node attributes.
func (m *MemFS) SetAttr(ino int64, attr *InoAttr, mask SetAttrMask, fi *FileInfo) (
	*InoAttr, Status,
) {
	slog.Debug("SetAttr", "ino", ino, "attr", attr, "mask", mask, "fi", fi)

	i := m.inodes[ino]
	if i == nil {
		return nil, ENOENT
	}

	if mask&SET_ATTR_MODE != 0 {
		i.mode = attr.Mode
	}
	if mask&SET_ATTR_MTIME != 0 {
		i.mtime = attr.MTime
	}
	if mask&SET_ATTR_MTIME_NOW != 0 {
		i.mtime = time.Now()
	}
	if mask&SET_ATTR_SIZE != 0 {
		if i.file == nil {
			return nil, EISDIR
		}
		if int(attr.Size) <= len(i.file.data) {
			i.file.data = i.file.data[:attr.Size]
		} else {
			data := make([]byte, attr.Size)
			copy(data, i.file.data)
			i.file.data = data
		}
	}

	return i.stat(), OK
}

// Lookup finds node by name.
func (m *MemFS) Lookup(parent int64, name string) (entry *Entry, err Status) {
	slog.Debug("Lookup", "parent", parent, "name", name)

	n, err := m.dirNode(parent)
	if err != OK {
		return nil, err
	}

	i, exist := n.dir.nodes[name]
	if !exist {
		return nil, ENOENT
	}
	node := m.inodes[i]
	return m.entry(node), OK
}

// StatFS returns filesystem stats.
func (m *MemFS) StatFS(ino int64) (stat *StatVFS, status Status) {
	slog.Debug("StatFS", "ino", ino)

	stat = &StatVFS{
		Files: int64(len(m.inodes)),
	}
	status = OK
	return
}

// Flush syncs filesystem data.
func (m *MemFS) Flush(ino int64, fi *FileInfo) Status {
	return OK
}

// ReadDir reads a directory.
func (m *MemFS) ReadDir(ino int64, fi *FileInfo, off int64, size int, w DirEntryWriter) Status {
	slog.Debug("ReadDir", "ino", ino, "off", off, "size", size)

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

// Open opens files.
func (m *MemFS) Open(ino int64, fi *FileInfo) Status {
	slog.Debug("Open", "ino", ino, "fi", fi)

	_, err := m.fileNode(ino)
	return err
}

// Rmdir removes directories.
func (m *MemFS) Rmdir(dir int64, name string) Status {
	slog.Debug("Rmdir", "dir", dir, "name", name)

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

// Rename changes names.
func (m *MemFS) Rename(dir int64, name string, newdir int64, newname string, flags int) Status {
	slog.Debug("Rename", "dir", dir, "name", name, "newdir", newdir, "newname", newname, "flags", flags)
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
	newOID, present := nd.dir.nodes[newname]
	if present {
		c := m.inodes[newOID]
		if c.file == nil {
			return EISDIR
		}

		delete(m.inodes, c.id)
	}

	nd.dir.nodes[newname] = oid
	delete(od.dir.nodes, name)
	return OK
}

// Unlink removes files.
func (m *MemFS) Unlink(dir int64, name string) Status {
	slog.Debug("Unlink", "dir", dir, "name", name)

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

// Read loads data from a file.
func (m *MemFS) Read(ino, size, off int64, fi *FileInfo) ([]byte, Status) {
	slog.Debug("Read", "ino", ino, "size", size, "off", off, "fi", fi)

	n, err := m.fileNode(ino)
	if err != OK {
		return nil, err
	}

	data := n.file.data
	avail := int64(len(data)) - off
	if avail < size {
		size = avail
	}
	if size <= 0 {
		return []byte{}, OK
	}
	return data[off : off+size], OK
}

// Write stores data to a file.
func (m *MemFS) Write(p []byte, ino int64, off int64, fi *FileInfo) (int, Status) {
	slog.Debug("Write", "ino", ino, "off", off, "fi", fi)

	n, err := m.fileNode(ino)
	if err != OK {
		return 0, err
	}

	if rl := int(off) + len(p); rl > len(n.file.data) {
		// Extend
		newSlice := make([]byte, rl)
		copy(newSlice, n.file.data)
		n.file.data = newSlice
	}
	slice := n.file.data[off:]
	copy(slice, p)
	return len(p), OK
}
