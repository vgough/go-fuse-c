package fuse

// #cgo pkg-config: fuse-t
// #cgo CFLAGS: -I/Library/Frameworks/fuse_t.framework/Headers
//
// #include "wrapper.h"
// #include <stdlib.h>  // for free()
import "C"

import "time"

func (s *StatVFS) toCStat(o *C.struct_statvfs) {
	o.f_bsize = C.ulong(s.BlockSize)
	o.f_blocks = C.fsblkcnt_t(s.Blocks)
	o.f_bfree = C.fsblkcnt_t(s.BlocksFree)

	o.f_files = C.fsfilcnt_t(s.Files)
	o.f_ffree = C.fsfilcnt_t(s.FilesFree)

	o.f_fsid = C.ulong(s.Fsid)
	o.f_flag = C.ulong(s.Flags)
	o.f_namemax = C.ulong(s.NameMax)
}

func (a *InoAttr) fromCStat(i *C.struct_stat) {
	a.Ino = int64(i.st_ino)
	a.Mode = int(i.st_mode)
	a.NLink = int(i.st_nlink)
	a.Size = int64(i.st_size)
	var uid int = int(i.st_uid)
	var gid int = int(i.st_gid)
	a.UID = &uid
	a.GID = &gid
	a.ATime = time.Unix(int64(i.st_atimespec.tv_sec), int64(i.st_atimespec.tv_nsec))
	a.CTime = time.Unix(int64(i.st_ctimespec.tv_sec), int64(i.st_ctimespec.tv_nsec))
	a.MTime = time.Unix(int64(i.st_mtimespec.tv_sec), int64(i.st_mtimespec.tv_nsec))
}

func (a *InoAttr) toCStat(o *C.struct_stat, timeout *C.double) {
	o.st_ino = C.__darwin_ino64_t(a.Ino)
	o.st_mode = C.mode_t(a.Mode)
	o.st_nlink = C.nlink_t(a.NLink)
	o.st_size = C.off_t(a.Size)
	if a.UID != nil {
		o.st_uid = C.uid_t(*a.UID)
	}
	if a.GID != nil {
		o.st_gid = C.gid_t(*a.GID)
	}
	toCTime(&o.st_ctimespec, a.CTime)
	toCTime(&o.st_mtimespec, a.MTime)
	toCTime(&o.st_atimespec, a.ATime)
	if timeout != nil {
		(*timeout) = C.double(a.Timeout)
	}
}
