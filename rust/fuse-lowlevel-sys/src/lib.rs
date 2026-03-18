#![allow(non_camel_case_types)]
#![allow(non_snake_case)]

use std::ffi::{c_char, c_int, c_uchar, c_uint, c_void};

pub type FuseLLReq = *mut c_void;

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLDuration {
    pub secs: u64,
    pub nanos: u32,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLStat {
    pub ino: u64,
    pub size: i64,
    pub mode: u32,
    pub nlink: u32,
    pub uid: u32,
    pub gid: u32,
    pub atime_secs: i64,
    pub atime_nanos: u32,
    pub ctime_secs: i64,
    pub ctime_nanos: u32,
    pub mtime_secs: i64,
    pub mtime_nanos: u32,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLStatVfs {
    pub block_size: u64,
    pub blocks: u64,
    pub blocks_free: u64,
    pub files: u64,
    pub files_free: u64,
    pub fsid: u64,
    pub flags: u64,
    pub name_max: u64,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLFileInfo {
    pub flags: c_int,
    pub writepage: c_int,
    pub handle: u64,
    pub lock_owner: u64,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLConnInfo {
    pub proto_major: c_int,
    pub proto_minor: c_int,
    pub max_write: c_uint,
    pub max_readahead: c_uint,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLRequestContext {
    pub uid: u32,
    pub gid: u32,
    pub pid: i32,
    pub umask: u32,
}

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLEntry {
    pub ino: u64,
    pub generation: u64,
    pub attr: FuseLLStat,
    pub attr_timeout: FuseLLDuration,
    pub entry_timeout: FuseLLDuration,
}

pub type FuseLLBufFreeFn =
    Option<unsafe extern "C" fn(data: *mut c_void, buf: *const c_uchar, len: usize)>;

#[repr(C)]
#[derive(Clone, Copy, Debug, Default)]
pub struct FuseLLBuf {
    pub data: *const c_uchar,
    pub len: usize,
    pub free_fn: FuseLLBufFreeFn,
    pub free_data: *mut c_void,
}

#[repr(C)]
pub struct FuseLLDirBuf {
    _private: [u8; 0],
}

#[repr(C)]
pub struct FuseLLSession {
    _private: [u8; 0],
}

pub type FuseLLInitFn =
    Option<unsafe extern "C" fn(userdata: *mut c_void, conn: *mut FuseLLConnInfo)>;
pub type FuseLLDestroyFn = Option<unsafe extern "C" fn(userdata: *mut c_void)>;
pub type FuseLLLookupFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;
pub type FuseLLForgetFn =
    Option<unsafe extern "C" fn(userdata: *mut c_void, req: FuseLLReq, ino: u64, nlookup: u64)>;
pub type FuseLLGetAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        fi: *const FuseLLFileInfo,
        out: *mut FuseLLStat,
        timeout: *mut FuseLLDuration,
    ) -> c_int,
>;
pub type FuseLLSetAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        attr: *const FuseLLStat,
        to_set: c_int,
        fi: *const FuseLLFileInfo,
        out: *mut FuseLLStat,
        timeout: *mut FuseLLDuration,
    ) -> c_int,
>;
pub type FuseLLReadLinkFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        out: *mut FuseLLBuf,
    ) -> c_int,
>;
pub type FuseLLNodeFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
        mode: c_uint,
        rdev: c_uint,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;
pub type FuseLLMkdirFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
        mode: c_uint,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;
pub type FuseLLUnlinkFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
    ) -> c_int,
>;
pub type FuseLLSymlinkFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        link: *const c_char,
        parent: u64,
        name: *const c_char,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;
pub type FuseLLRenameFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
        newparent: u64,
        newname: *const c_char,
        flags: c_uint,
    ) -> c_int,
>;
pub type FuseLLLinkFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        newparent: u64,
        name: *const c_char,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;
pub type FuseLLOpenFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        fi: *mut FuseLLFileInfo,
    ) -> c_int,
>;
pub type FuseLLReadFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        size: usize,
        off: i64,
        fi: *const FuseLLFileInfo,
        out: *mut FuseLLBuf,
    ) -> c_int,
>;
pub type FuseLLWriteFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        buf: *const c_uchar,
        size: usize,
        off: i64,
        fi: *const FuseLLFileInfo,
        written: *mut usize,
    ) -> c_int,
>;
pub type FuseLLFileOnlyFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        fi: *const FuseLLFileInfo,
    ) -> c_int,
>;
pub type FuseLLFsyncFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        datasync: c_int,
        fi: *const FuseLLFileInfo,
    ) -> c_int,
>;
pub type FuseLLReadDirFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        fi: *const FuseLLFileInfo,
        off: i64,
        size: usize,
        db: *mut FuseLLDirBuf,
    ) -> c_int,
>;
pub type FuseLLStatFsFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        out: *mut FuseLLStatVfs,
    ) -> c_int,
>;
pub type FuseLLSetXAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        name: *const c_char,
        value: *const c_uchar,
        size: usize,
        flags: c_int,
    ) -> c_int,
>;
pub type FuseLLGetXAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        name: *const c_char,
        out: *mut FuseLLBuf,
    ) -> c_int,
>;
pub type FuseLLListXAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        out: *mut FuseLLBuf,
    ) -> c_int,
>;
pub type FuseLLRemoveXAttrFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        ino: u64,
        name: *const c_char,
    ) -> c_int,
>;
pub type FuseLLAccessFn = Option<
    unsafe extern "C" fn(userdata: *mut c_void, req: FuseLLReq, ino: u64, mask: c_int) -> c_int,
>;
pub type FuseLLCreateFn = Option<
    unsafe extern "C" fn(
        userdata: *mut c_void,
        req: FuseLLReq,
        parent: u64,
        name: *const c_char,
        mode: c_uint,
        fi: *mut FuseLLFileInfo,
        out: *mut FuseLLEntry,
    ) -> c_int,
>;

#[repr(C)]
#[derive(Clone, Copy)]
pub struct FuseLLOps {
    pub userdata: *mut c_void,
    pub init: FuseLLInitFn,
    pub destroy: FuseLLDestroyFn,
    pub lookup: FuseLLLookupFn,
    pub forget: FuseLLForgetFn,
    pub getattr: FuseLLGetAttrFn,
    pub setattr: FuseLLSetAttrFn,
    pub readlink: FuseLLReadLinkFn,
    pub mknod: FuseLLNodeFn,
    pub mkdir: FuseLLMkdirFn,
    pub unlink: FuseLLUnlinkFn,
    pub rmdir: FuseLLUnlinkFn,
    pub symlink: FuseLLSymlinkFn,
    pub rename: FuseLLRenameFn,
    pub link: FuseLLLinkFn,
    pub open: FuseLLOpenFn,
    pub read: FuseLLReadFn,
    pub write: FuseLLWriteFn,
    pub flush: FuseLLFileOnlyFn,
    pub release: FuseLLFileOnlyFn,
    pub fsync: FuseLLFsyncFn,
    pub opendir: FuseLLOpenFn,
    pub readdir: FuseLLReadDirFn,
    pub releasedir: FuseLLFileOnlyFn,
    pub fsyncdir: FuseLLFsyncFn,
    pub statfs: FuseLLStatFsFn,
    pub setxattr: FuseLLSetXAttrFn,
    pub getxattr: FuseLLGetXAttrFn,
    pub listxattr: FuseLLListXAttrFn,
    pub removexattr: FuseLLRemoveXAttrFn,
    pub access: FuseLLAccessFn,
    pub create: FuseLLCreateFn,
}

unsafe extern "C" {
    pub fn fuse_ll_version() -> c_int;
    pub fn fuse_ll_session_mount(
        argc: c_int,
        argv: *const *const c_char,
        ops: *const FuseLLOps,
    ) -> *mut FuseLLSession;
    pub fn fuse_ll_session_run(session: *mut FuseLLSession) -> c_int;
    pub fn fuse_ll_session_unmount(session: *mut FuseLLSession);
    pub fn fuse_ll_session_destroy(session: *mut FuseLLSession);
    pub fn fuse_ll_req_ctx(req: FuseLLReq, out: *mut FuseLLRequestContext) -> c_int;
    pub fn fuse_ll_dirbuf_add(
        db: *mut FuseLLDirBuf,
        name: *const c_char,
        ino: u64,
        mode: c_uint,
        next: i64,
    ) -> c_int;
}
