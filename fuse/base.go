package fuse

type DefaultRawFileSystem struct {
}

func (d *DefaultRawFileSystem) Init(*ConnInfo) {}

func (d *DefaultRawFileSystem) Destroy() {}
func (d *DefaultRawFileSystem) StatFs(ino int64, stat *StatVfs) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Lookup(dir int64, name string) (entry *EntryParam, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Forget(ino int64, n int) {}

func (d *DefaultRawFileSystem) GetAttr(ino int64, fi *FileInfo) (attr *InoAttr, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) ReadDir(ino int64, fi *FileInfo, off int64, size int,
	w DirEntryWriter) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Mknod(p int64, name string, mode int, rdev int) (
	entry *EntryParam, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Open(ino int64, fi *FileInfo) Status {
	return ENOSYS
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
	entry *EntryParam, err Status) {
	return nil, ENOSYS
}

func (d *DefaultRawFileSystem) Rmdir(p int64, name string) Status {
	return ENOSYS
}

func (d *DefaultRawFileSystem) Rename(int64, string, int64, string) Status {
	return ENOSYS
}
