package fuse

// DefaultRawFileSystem provides a filesystem that returns a suitable default for all methods.
// Most methods allow ENOSYS, which signals to FUSE that the operation is not implemented.
// Other methods simply return success, if the method is optional.
//
// This implementation is intended to be used as the base implementation for a filesystem, so that
// all methods not implemented by the derived type will be handled here.
//
// Usage example:
//   type MyFs struct {
//     fuse.DefaultRawFileSystem
//   }
type DefaultRawFileSystem struct {
}

func (d *DefaultRawFileSystem) Init(*ConnInfo) {}

func (d *DefaultRawFileSystem) Destroy() {}
func (d *DefaultRawFileSystem) StatFs(ino int64) (*StatVfs, Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Lookup(dir int64, name string) (entry *Entry, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Forget(ino int64, n int) {}

func (d *DefaultRawFileSystem) Release(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) ReleaseDir(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) FSync(ino int64, dataOnly bool, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) FSyncDir(ino int64, dataOnly bool, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Flush(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) GetAttr(ino int64, fi *FileInfo) (attr *InoAttr, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) SetAttr(ino int64, attr *InoAttr, mask SetAttrMask, fi *FileInfo) (
	*InoAttr, Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) ReadLink(ino int64) (string, Status) {
	return "", ENOSYS
}

func (d *DefaultRawFileSystem) ReadDir(ino int64, fi *FileInfo, off int64, size int,
	w DirEntryWriter) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Mknod(p int64, name string, mode int, rdev int) (
	entry *Entry, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Access(ino int64, mode int) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Create(p int64, name string, mode int, fi *FileInfo) (
	entry *Entry, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Open(ino int64, fi *FileInfo) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) OpenDir(ino int64, fi *FileInfo) Status {
	return OK
}

func (d *DefaultRawFileSystem) Read(p []byte, ino int64, off int64, fi *FileInfo) (
	n int, err Status) {
	return 0, ENOSYS
}

func (d *DefaultRawFileSystem) Write(p []byte, ino int64, off int64, fi *FileInfo) (
	n int, err Status) {
	return 0, ENOSYS
}

func (d *DefaultRawFileSystem) Mkdir(p int64, name string, mode int) (
	entry *Entry, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Rmdir(p int64, name string) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Symlink(link string, p int64, name string) (*Entry, Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Link(ino int64, newparent int64, name string) (*Entry, Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Rename(int64, string, int64, string) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Unlink(p int64, name string) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) ListXattrs(ino int64) ([]string, Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) GetXattrSize(ino int64, name string) (int, Status) {
	return 0, ENOSYS
}

func (d *DefaultRawFileSystem) GetXattr(ino int64, name string, out []byte) (int, Status) {
	return 0, ENOSYS
}

func (d *DefaultRawFileSystem) SetXattr(ino int64, name string, value []byte, flags int) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) RemoveXattr(ino int64, name string) Status {
	return ENOSYS
}
