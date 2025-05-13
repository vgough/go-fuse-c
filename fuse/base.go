package fuse

// DefaultFileSystem provides a filesystem that returns a suitable default for all methods.
// Most methods allow ENOSYS, which signals to FUSE that the operation is not implemented.
// Other methods simply return success, if the method is optional.
//
// This implementation is intended to be used as the base implementation for a filesystem, so that
// all methods not implemented by the derived type will be handled here.
//
// Usage eXAmple:
//
//	type MyFs struct {
//	  fuse.DefaultFileSystem
//	}
type DefaultFileSystem struct{}

var _ FileSystem = &DefaultFileSystem{}

// Init implements FileSystem.
func (d *DefaultFileSystem) Init(*ConnInfo) {}

// Destroy implements FileSystem.
func (d *DefaultFileSystem) Destroy() {}

// StatFS implements FileSystem.
func (d *DefaultFileSystem) StatFS(ino int64) (*StatVFS, Status) {
	return nil, ENOSYS
}

// Lookup implements FileSystem.
func (d *DefaultFileSystem) Lookup(dir int64, name string) (entry *Entry, err Status) {
	return nil, ENOSYS
}

// Forget implements FileSystem.
func (d *DefaultFileSystem) Forget(ino int64, n int) {}

// Release implements FileSystem.
func (d *DefaultFileSystem) Release(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

// ReleaseDir implements FileSystem.
func (d *DefaultFileSystem) ReleaseDir(ino int64, fi *FileInfo) Status {
	return 0
}

// FSync implements FileSystem.
func (d *DefaultFileSystem) FSync(ino int64, dataOnly bool, fi *FileInfo) Status {
	return ENOSYS
}

// FSyncDir implements FileSystem.
func (d *DefaultFileSystem) FSyncDir(ino int64, dataOnly bool, fi *FileInfo) Status {
	return ENOSYS
}

// Flush implements FileSystem.
func (d *DefaultFileSystem) Flush(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

// GetAttr implements FileSystem.
func (d *DefaultFileSystem) GetAttr(ino int64, fi *FileInfo) (attr *InoAttr, err Status) {
	return nil, ENOSYS
}

// SetAttr implements FileSystem.
func (d *DefaultFileSystem) SetAttr(ino int64, attr *InoAttr, mask SetAttrMask, fi *FileInfo) (
	*InoAttr, Status,
) {
	return nil, ENOSYS
}

// ReadLink implements FileSystem.
func (d *DefaultFileSystem) ReadLink(ino int64) (string, Status) {
	return "", ENOSYS
}

// ReadDir implements FileSystem.
func (d *DefaultFileSystem) ReadDir(ino int64, fi *FileInfo, off int64, size int,
	w DirEntryWriter,
) Status {
	return ENOSYS
}

// Mknod implements FileSystem.
func (d *DefaultFileSystem) Mknod(p int64, name string, mode int, rdev int) (
	entry *Entry, err Status,
) {
	return nil, ENOSYS
}

// Access implements FileSystem.
func (d *DefaultFileSystem) Access(ino int64, mode int) Status {
	return ENOSYS
}

// Create implements FileSystem.
func (d *DefaultFileSystem) Create(p int64, name string, mode int, fi *FileInfo) (
	entry *Entry, err Status,
) {
	return nil, ENOSYS
}

// Open implements FileSystem.
func (d *DefaultFileSystem) Open(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

// OpenDir implements FileSystem.
func (d *DefaultFileSystem) OpenDir(ino int64, fi *FileInfo) Status {
	return OK
}

// Read implements FileSystem.
func (d *DefaultFileSystem) Read(ino int64, size int64, off int64, fi *FileInfo) (
	data []byte, err Status,
) {
	return nil, ENOSYS
}

// Write implements FileSystem.
func (d *DefaultFileSystem) Write(p []byte, ino int64, off int64, fi *FileInfo) (
	n int, err Status,
) {
	return 0, ENOSYS
}

// Mkdir implements FileSystem.
func (d *DefaultFileSystem) Mkdir(p int64, name string, mode int) (
	entry *Entry, err Status,
) {
	return nil, ENOSYS
}

// Rmdir implements FileSystem.
func (d *DefaultFileSystem) Rmdir(p int64, name string) Status {
	return ENOSYS
}

// Symlink implements FileSystem.
func (d *DefaultFileSystem) Symlink(link string, p int64, name string) (*Entry, Status) {
	return nil, ENOSYS
}

// Link implements FileSystem.
func (d *DefaultFileSystem) Link(ino int64, newparent int64, name string) (*Entry, Status) {
	return nil, ENOSYS
}

// Rename implements FileSystem.
func (d *DefaultFileSystem) Rename(int64, string, int64, string, int) Status {
	return ENOSYS
}

// Unlink implements FileSystem.
func (d *DefaultFileSystem) Unlink(p int64, name string) Status {
	return ENOSYS
}

// ListXAttrs implements FileSystem.
func (d *DefaultFileSystem) ListXAttrs(ino int64) ([]string, Status) {
	return nil, ENOSYS
}

// GetXAttrSize implements FileSystem.
func (d *DefaultFileSystem) GetXAttrSize(ino int64, name string) (int, Status) {
	return 0, ENOSYS
}

// GetXAttr implements FileSystem.
func (d *DefaultFileSystem) GetXAttr(ino int64, name string, out []byte) (int, Status) {
	return 0, ENOSYS
}

// SetXAttr implements FileSystem.
func (d *DefaultFileSystem) SetXAttr(ino int64, name string, value []byte, flags int) Status {
	return ENOSYS
}

// RemoveXAttr implements FileSystem.
func (d *DefaultFileSystem) RemoveXAttr(ino int64, name string) Status {
	return ENOSYS
}
