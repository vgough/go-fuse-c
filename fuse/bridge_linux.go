package fuse

// #cgo pkg-config: fuse3
//
// #include "wrapper.h"
// #include <stdlib.h>  // for free()
import "C"

import "time"

func (s *StatVFS) toCStat(o *C.struct_statvfs) {
	o.f_bsize = C.ulong(s.BlockSize)
	o.f_blocks = C.__fsblkcnt64_t(s.Blocks)
	o.f_bfree = C.__fsblkcnt64_t(s.BlocksFree)

	o.f_files = C.__fsfilcnt64_t(s.Files)
	o.f_ffree = C.__fsfilcnt64_t(s.FilesFree)

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
	a.ATime = time.Unix(int64(i.st_atim.tv_sec), int64(i.st_atim.tv_nsec))
	a.CTime = time.Unix(int64(i.st_ctim.tv_sec), int64(i.st_ctim.tv_nsec))
	a.MTime = time.Unix(int64(i.st_mtim.tv_sec), int64(i.st_mtim.tv_nsec))
}

func (a *InoAttr) toCStat(o *C.struct_stat, timeout *C.double) {
	o.st_ino = C.__ino_t(a.Ino)
	o.st_mode = C.__mode_t(a.Mode)
	o.st_nlink = C.__nlink_t(a.NLink)
	o.st_size = C.__off_t(a.Size)
	if a.UID != nil {
		o.st_uid = C.__uid_t(*a.UID)
	}
	if a.GID != nil {
		o.st_gid = C.__gid_t(*a.GID)
	}

	toCTime(&o.st_ctim, a.CTime)
	toCTime(&o.st_mtim, a.MTime)
	toCTime(&o.st_atim, a.ATime)
	if timeout != nil {
		(*timeout) = C.double(a.Timeout)
	}
}
