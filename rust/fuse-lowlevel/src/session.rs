use crate::{
    ConnInfo, DirEntryWriter, Entry, Errno, FUSE_ROOT_ID, FileInfo, Filesystem, FuseResult, Handle,
    MountError, NodeAttr, RequestContext, S_IFDIR, SetAttrMask, StatFs,
};
use fuse_lowlevel_sys as sys;
use std::ffi::{CStr, CString, OsStr, OsString, c_char, c_int, c_uchar, c_void};
use std::marker::PhantomData;
use std::mem;
use std::os::unix::ffi::{OsStrExt, OsStringExt};
use std::panic::{AssertUnwindSafe, catch_unwind};
use std::ptr;
use std::slice;
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

/// Returns the version reported by the linked native FUSE library.
pub fn version() -> i32 {
    unsafe { sys::fuse_ll_version() }
}

/// Typed options used when creating a mounted FUSE session.
#[derive(Clone, Debug)]
pub struct MountOptions {
    program_name: OsString,
    foreground: bool,
    single_thread: bool,
    raw_args: Vec<OsString>,
}

impl Default for MountOptions {
    fn default() -> Self {
        Self {
            program_name: OsString::from("fuse-lowlevel"),
            foreground: false,
            single_thread: false,
            raw_args: Vec::new(),
        }
    }
}

impl MountOptions {
    /// Creates an empty set of mount options.
    pub fn new() -> Self {
        Self::default()
    }

    /// Sets the program name forwarded to libfuse.
    pub fn program_name(mut self, value: impl Into<OsString>) -> Self {
        self.program_name = value.into();
        self
    }

    /// Enables foreground mode (`-f`) for the mounted session.
    pub fn foreground(mut self, enabled: bool) -> Self {
        self.foreground = enabled;
        self
    }

    /// Enables single-threaded request handling (`-s`).
    pub fn single_thread(mut self, enabled: bool) -> Self {
        self.single_thread = enabled;
        self
    }

    /// Appends a raw libfuse argument that is passed through unchanged.
    pub fn raw_arg(mut self, value: impl Into<OsString>) -> Self {
        self.raw_args.push(value.into());
        self
    }

    fn argv(&self, mountpoint: &OsStr) -> Result<Vec<CString>, MountError> {
        let mut args = Vec::with_capacity(4 + self.raw_args.len());
        args.push(to_cstring(self.program_name.as_os_str())?);
        if self.foreground {
            args.push(CString::new("-f").map_err(|_| MountError::NulByte)?);
        }
        if self.single_thread {
            args.push(CString::new("-s").map_err(|_| MountError::NulByte)?);
        }
        for raw in &self.raw_args {
            args.push(to_cstring(raw.as_os_str())?);
        }
        args.push(to_cstring(mountpoint)?);
        Ok(args)
    }
}

struct SessionState<F: Filesystem> {
    fs: Arc<F>,
}

/// Mounted low-level FUSE session.
///
/// `Session` owns the native session handle and unmounts/destroys it on drop.
pub struct Session<F: Filesystem> {
    state: *mut SessionState<F>,
    raw: *mut sys::FuseLLSession,
    _marker: PhantomData<F>,
}

impl<F: Filesystem> Session<F> {
    /// Mounts a filesystem using an owned filesystem value.
    pub fn mount(
        filesystem: F,
        mountpoint: impl AsRef<OsStr>,
        options: MountOptions,
    ) -> Result<Self, MountError> {
        Self::mount_arc(Arc::new(filesystem), mountpoint.as_ref(), options)
    }

    /// Mounts a filesystem using a shared [`Arc`] instance.
    pub fn mount_arc(
        filesystem: Arc<F>,
        mountpoint: &OsStr,
        options: MountOptions,
    ) -> Result<Self, MountError> {
        let mut argv = options.argv(mountpoint)?;
        let raw_argv: Vec<*const c_char> = argv.iter().map(|arg| arg.as_ptr()).collect();

        let state = Box::new(SessionState { fs: filesystem });
        let state_ptr = Box::into_raw(state);
        let ops = build_ops::<F>(state_ptr.cast::<c_void>());
        let session =
            unsafe { sys::fuse_ll_session_mount(raw_argv.len() as c_int, raw_argv.as_ptr(), &ops) };

        if session.is_null() {
            unsafe {
                drop(Box::from_raw(state_ptr));
            }
            return Err(MountError::MountFailed);
        }

        argv.clear();

        Ok(Self {
            state: state_ptr,
            raw: session,
            _marker: PhantomData,
        })
    }

    /// Runs the FUSE event loop until the session exits.
    pub fn run(&mut self) -> Result<(), MountError> {
        let status = unsafe { sys::fuse_ll_session_run(self.raw) };
        if status == 0 {
            Ok(())
        } else {
            Err(MountError::RunFailed(status))
        }
    }

    /// Requests that the mounted session be unmounted.
    ///
    /// This is optional because the session is also unmounted during [`Drop`].
    pub fn unmount(&mut self) {
        if !self.raw.is_null() {
            unsafe { sys::fuse_ll_session_unmount(self.raw) };
        }
    }

    /// Returns a clone of the filesystem object owned by the session.
    pub fn filesystem(&self) -> Arc<F> {
        unsafe { (*self.state).fs.clone() }
    }
}

impl<F: Filesystem> Drop for Session<F> {
    fn drop(&mut self) {
        if !self.raw.is_null() {
            unsafe {
                sys::fuse_ll_session_destroy(self.raw);
            }
            self.raw = ptr::null_mut();
        }
        if !self.state.is_null() {
            unsafe {
                drop(Box::from_raw(self.state));
            }
            self.state = ptr::null_mut();
        }
    }
}

/// Convenience helper that mounts a filesystem and immediately runs the FUSE
/// event loop.
pub fn mount_and_run<F: Filesystem>(
    filesystem: F,
    mountpoint: impl AsRef<OsStr>,
    options: MountOptions,
) -> Result<(), MountError> {
    let mut session = Session::mount(filesystem, mountpoint, options)?;
    session.run()
}

struct SysDirWriter {
    raw: *mut sys::FuseLLDirBuf,
}

impl DirEntryWriter for SysDirWriter {
    fn add(&mut self, name: &OsStr, ino: u64, mode: u32, next: i64) -> bool {
        let c_name = match CString::new(name.as_bytes()) {
            Ok(name) => name,
            Err(_) => return false,
        };
        unsafe { sys::fuse_ll_dirbuf_add(self.raw, c_name.as_ptr(), ino, mode, next) == 0 }
    }
}

unsafe extern "C" fn free_vec_buffer(data: *mut c_void, buf: *const c_uchar, len: usize) {
    let _ = data;
    if !buf.is_null() {
        let ptr = buf as *mut u8;
        unsafe {
            drop(Vec::from_raw_parts(ptr, len, len));
        }
    }
}

fn into_sys_buf(bytes: Vec<u8>) -> sys::FuseLLBuf {
    let mut bytes = bytes;
    let data = bytes.as_mut_ptr();
    let len = bytes.len();
    mem::forget(bytes);
    sys::FuseLLBuf {
        data,
        len,
        free_fn: Some(free_vec_buffer),
        free_data: ptr::null_mut(),
    }
}

fn listxattrs_to_buf(values: Vec<OsString>) -> Vec<u8> {
    let mut out = Vec::new();
    for value in values {
        out.extend_from_slice(&value.into_vec());
        out.push(0);
    }
    out
}

fn to_cstring(value: &OsStr) -> Result<CString, MountError> {
    CString::new(value.as_bytes()).map_err(|_| MountError::NulByte)
}

fn from_c_name<'a>(value: *const c_char) -> &'a OsStr {
    let bytes = unsafe { CStr::from_ptr(value) }.to_bytes();
    OsStr::from_bytes(bytes)
}

fn request_context(req: sys::FuseLLReq) -> RequestContext {
    let mut raw = sys::FuseLLRequestContext::default();
    let _ = unsafe { sys::fuse_ll_req_ctx(req, &mut raw) };
    RequestContext {
        uid: raw.uid,
        gid: raw.gid,
        pid: raw.pid,
        umask: raw.umask,
    }
}

fn system_time_from_parts(secs: i64, nanos: u32) -> SystemTime {
    if secs >= 0 {
        UNIX_EPOCH + Duration::new(secs as u64, nanos)
    } else {
        UNIX_EPOCH - Duration::new(secs.unsigned_abs(), nanos)
    }
}

fn duration_to_sys(value: Duration) -> sys::FuseLLDuration {
    sys::FuseLLDuration {
        secs: value.as_secs(),
        nanos: value.subsec_nanos(),
    }
}

fn node_attr_to_sys(attr: &NodeAttr) -> sys::FuseLLStat {
    let atime = attr
        .atime
        .duration_since(UNIX_EPOCH)
        .unwrap_or_else(|_| Duration::from_secs(0));
    let ctime = attr
        .ctime
        .duration_since(UNIX_EPOCH)
        .unwrap_or_else(|_| Duration::from_secs(0));
    let mtime = attr
        .mtime
        .duration_since(UNIX_EPOCH)
        .unwrap_or_else(|_| Duration::from_secs(0));

    sys::FuseLLStat {
        ino: attr.ino,
        size: attr.size,
        mode: attr.mode,
        nlink: attr.nlink,
        uid: attr.uid,
        gid: attr.gid,
        atime_secs: atime.as_secs() as i64,
        atime_nanos: atime.subsec_nanos(),
        ctime_secs: ctime.as_secs() as i64,
        ctime_nanos: ctime.subsec_nanos(),
        mtime_secs: mtime.as_secs() as i64,
        mtime_nanos: mtime.subsec_nanos(),
    }
}

fn normalize_attr_reply(attr: &mut sys::FuseLLStat, fallback_ino: u64) {
    const S_IFMT: u32 = 0o170000;

    if attr.ino == 0 && fallback_ino != 0 {
        attr.ino = fallback_ino;
    }

    if attr.ino == FUSE_ROOT_ID && (attr.mode & S_IFMT) == 0 {
        attr.mode |= S_IFDIR;
    }

    if attr.nlink == 0 {
        attr.nlink = if (attr.mode & S_IFMT) == S_IFDIR {
            2
        } else {
            1
        };
    }
}

fn node_attr_from_sys(attr: &sys::FuseLLStat, timeout: Duration) -> NodeAttr {
    NodeAttr {
        ino: attr.ino,
        size: attr.size,
        mode: attr.mode,
        nlink: attr.nlink,
        uid: attr.uid,
        gid: attr.gid,
        atime: system_time_from_parts(attr.atime_secs, attr.atime_nanos),
        ctime: system_time_from_parts(attr.ctime_secs, attr.ctime_nanos),
        mtime: system_time_from_parts(attr.mtime_secs, attr.mtime_nanos),
        timeout,
    }
}

fn file_info_from_sys(info: &sys::FuseLLFileInfo) -> FileInfo {
    FileInfo {
        flags: info.flags,
        writepage: info.writepage != 0,
        handle: Handle(info.handle),
        lock_owner: info.lock_owner,
    }
}

fn file_info_to_sys(info: &FileInfo, out: &mut sys::FuseLLFileInfo) {
    out.flags = info.flags;
    out.writepage = if info.writepage { 1 } else { 0 };
    out.handle = info.handle.0;
    out.lock_owner = info.lock_owner;
}

fn conn_info_from_sys(info: &sys::FuseLLConnInfo) -> ConnInfo {
    ConnInfo {
        proto_major: info.proto_major,
        proto_minor: info.proto_minor,
        max_write: info.max_write,
        max_readahead: info.max_readahead,
    }
}

fn conn_info_to_sys(info: &ConnInfo, out: &mut sys::FuseLLConnInfo) {
    out.proto_major = info.proto_major;
    out.proto_minor = info.proto_minor;
    out.max_write = info.max_write;
    out.max_readahead = info.max_readahead;
}

fn entry_to_sys(entry: &Entry, out: &mut sys::FuseLLEntry) {
    out.ino = entry.ino;
    out.generation = if entry.generation == 0 {
        1
    } else {
        entry.generation
    };
    out.attr = node_attr_to_sys(&entry.attr);
    normalize_attr_reply(&mut out.attr, entry.ino);
    out.attr_timeout = duration_to_sys(entry.attr_timeout);
    out.entry_timeout = duration_to_sys(entry.entry_timeout);
}

fn statfs_to_sys(stat: &StatFs, out: &mut sys::FuseLLStatVfs) {
    out.block_size = stat.block_size;
    out.blocks = stat.blocks;
    out.blocks_free = stat.blocks_free;
    out.files = stat.files;
    out.files_free = stat.files_free;
    out.fsid = stat.fsid;
    out.flags = stat.flags.0;
    out.name_max = stat.name_max;
}

fn normalize_statfs_reply(stat: &mut sys::FuseLLStatVfs) {
    const DEFAULT_BLOCK_SIZE: u64 = 4096;
    const DEFAULT_NAME_MAX: u64 = 255;

    if stat.block_size == 0 {
        stat.block_size = DEFAULT_BLOCK_SIZE;
    }

    if stat.name_max == 0 {
        stat.name_max = DEFAULT_NAME_MAX;
    }
}

unsafe fn state_ref<'a, F: Filesystem>(userdata: *mut c_void) -> &'a SessionState<F> {
    unsafe { &*(userdata as *const SessionState<F>) }
}

fn with_errno<F>(f: F) -> c_int
where
    F: FnOnce() -> FuseResult<()>,
{
    match catch_unwind(AssertUnwindSafe(f)) {
        Ok(Ok(())) => 0,
        Ok(Err(err)) => err.raw(),
        Err(_) => Errno::EIO.raw(),
    }
}

unsafe extern "C" fn op_init<F: Filesystem>(userdata: *mut c_void, conn: *mut sys::FuseLLConnInfo) {
    let _ = catch_unwind(AssertUnwindSafe(|| unsafe {
        let state = state_ref::<F>(userdata);
        let mut conn_info = conn_info_from_sys(&*conn);
        state.fs.init(&mut conn_info);
        conn_info_to_sys(&conn_info, &mut *conn);
    }));
}

unsafe extern "C" fn op_destroy<F: Filesystem>(userdata: *mut c_void) {
    let _ = catch_unwind(AssertUnwindSafe(|| unsafe {
        let state = state_ref::<F>(userdata);
        state.fs.destroy();
    }));
}

unsafe extern "C" fn op_lookup<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let entry = state.fs.lookup(&ctx, parent, from_c_name(name))?;
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_forget<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    nlookup: u64,
) {
    let _ = catch_unwind(AssertUnwindSafe(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.forget(&ctx, ino, nlookup);
    }));
}

unsafe extern "C" fn op_getattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *const sys::FuseLLFileInfo,
    out: *mut sys::FuseLLStat,
    timeout: *mut sys::FuseLLDuration,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = fi.as_ref().map(file_info_from_sys);
        let attr = state.fs.getattr(&ctx, ino, file_info.as_ref())?;
        *out = node_attr_to_sys(&attr);
        normalize_attr_reply(&mut *out, ino);
        *timeout = duration_to_sys(attr.timeout);
        Ok(())
    })
}

unsafe extern "C" fn op_setattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    attr: *const sys::FuseLLStat,
    to_set: c_int,
    fi: *const sys::FuseLLFileInfo,
    out: *mut sys::FuseLLStat,
    timeout: *mut sys::FuseLLDuration,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let input = node_attr_from_sys(&*attr, Duration::from_secs(0));
        let file_info = fi.as_ref().map(file_info_from_sys);
        let attr = state.fs.setattr(
            &ctx,
            ino,
            &input,
            SetAttrMask(to_set as u32),
            file_info.as_ref(),
        )?;
        *out = node_attr_to_sys(&attr);
        normalize_attr_reply(&mut *out, ino);
        *timeout = duration_to_sys(attr.timeout);
        Ok(())
    })
}

unsafe extern "C" fn op_readlink<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    out: *mut sys::FuseLLBuf,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let target = state.fs.readlink(&ctx, ino)?;
        *out = into_sys_buf(target.into_vec());
        Ok(())
    })
}

unsafe extern "C" fn op_mknod<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
    mode: u32,
    rdev: u32,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let entry = state
            .fs
            .mknod(&ctx, parent, from_c_name(name), mode, rdev)?;
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_mkdir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
    mode: u32,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let entry = state.fs.mkdir(&ctx, parent, from_c_name(name), mode)?;
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_unlink<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.unlink(&ctx, parent, from_c_name(name))
    })
}

unsafe extern "C" fn op_rmdir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.rmdir(&ctx, parent, from_c_name(name))
    })
}

unsafe extern "C" fn op_symlink<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    link: *const c_char,
    parent: u64,
    name: *const c_char,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let entry = state
            .fs
            .symlink(&ctx, from_c_name(link), parent, from_c_name(name))?;
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_rename<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
    newparent: u64,
    newname: *const c_char,
    flags: u32,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.rename(
            &ctx,
            parent,
            from_c_name(name),
            newparent,
            from_c_name(newname),
            flags,
        )
    })
}

unsafe extern "C" fn op_link<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    newparent: u64,
    name: *const c_char,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let entry = state.fs.link(&ctx, ino, newparent, from_c_name(name))?;
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_open<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *mut sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let mut file_info = file_info_from_sys(&*fi);
        state.fs.open(&ctx, ino, &mut file_info)?;
        file_info_to_sys(&file_info, &mut *fi);
        Ok(())
    })
}

unsafe extern "C" fn op_read<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    size: usize,
    off: i64,
    fi: *const sys::FuseLLFileInfo,
    out: *mut sys::FuseLLBuf,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        let data = state.fs.read(&ctx, ino, size, off, &file_info)?;
        *out = into_sys_buf(data);
        Ok(())
    })
}

unsafe extern "C" fn op_write<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    buf: *const c_uchar,
    size: usize,
    off: i64,
    fi: *const sys::FuseLLFileInfo,
    written: *mut usize,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        let data = slice::from_raw_parts(buf, size);
        *written = state.fs.write(&ctx, data, ino, off, &file_info)?;
        Ok(())
    })
}

unsafe extern "C" fn op_flush<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *const sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        state.fs.flush(&ctx, ino, &file_info)
    })
}

unsafe extern "C" fn op_release<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *const sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        state.fs.release(&ctx, ino, &file_info)
    })
}

unsafe extern "C" fn op_fsync<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    datasync: c_int,
    fi: *const sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        state.fs.fsync(&ctx, ino, datasync != 0, &file_info)
    })
}

unsafe extern "C" fn op_opendir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *mut sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let mut file_info = file_info_from_sys(&*fi);
        state.fs.opendir(&ctx, ino, &mut file_info)?;
        file_info_to_sys(&file_info, &mut *fi);
        Ok(())
    })
}

unsafe extern "C" fn op_readdir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *const sys::FuseLLFileInfo,
    off: i64,
    size: usize,
    db: *mut sys::FuseLLDirBuf,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        let mut writer = SysDirWriter { raw: db };
        state
            .fs
            .readdir(&ctx, ino, &file_info, off, size, &mut writer)
    })
}

unsafe extern "C" fn op_releasedir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    fi: *const sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        state.fs.releasedir(&ctx, ino, &file_info)
    })
}

unsafe extern "C" fn op_fsyncdir<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    datasync: c_int,
    fi: *const sys::FuseLLFileInfo,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let file_info = file_info_from_sys(&*fi);
        state.fs.fsyncdir(&ctx, ino, datasync != 0, &file_info)
    })
}

unsafe extern "C" fn op_statfs<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    out: *mut sys::FuseLLStatVfs,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let stat = state.fs.statfs(&ctx, ino)?;
        statfs_to_sys(&stat, &mut *out);
        normalize_statfs_reply(&mut *out);
        Ok(())
    })
}

unsafe extern "C" fn op_setxattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    name: *const c_char,
    value: *const c_uchar,
    size: usize,
    flags: c_int,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let bytes = slice::from_raw_parts(value, size);
        state
            .fs
            .setxattr(&ctx, ino, from_c_name(name), bytes, flags)
    })
}

unsafe extern "C" fn op_getxattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    name: *const c_char,
    out: *mut sys::FuseLLBuf,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let value = state.fs.getxattr(&ctx, ino, from_c_name(name))?;
        *out = into_sys_buf(value);
        Ok(())
    })
}

unsafe extern "C" fn op_listxattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    out: *mut sys::FuseLLBuf,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let values = state.fs.listxattrs(&ctx, ino)?;
        *out = into_sys_buf(listxattrs_to_buf(values));
        Ok(())
    })
}

unsafe extern "C" fn op_removexattr<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    name: *const c_char,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.removexattr(&ctx, ino, from_c_name(name))
    })
}

unsafe extern "C" fn op_access<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    ino: u64,
    mask: c_int,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        state.fs.access(&ctx, ino, mask)
    })
}

unsafe extern "C" fn op_create<F: Filesystem>(
    userdata: *mut c_void,
    req: sys::FuseLLReq,
    parent: u64,
    name: *const c_char,
    mode: u32,
    fi: *mut sys::FuseLLFileInfo,
    out: *mut sys::FuseLLEntry,
) -> c_int {
    with_errno(|| unsafe {
        let state = state_ref::<F>(userdata);
        let ctx = request_context(req);
        let mut file_info = file_info_from_sys(&*fi);
        let entry = state
            .fs
            .create(&ctx, parent, from_c_name(name), mode, &mut file_info)?;
        file_info_to_sys(&file_info, &mut *fi);
        entry_to_sys(&entry, &mut *out);
        Ok(())
    })
}

fn build_ops<F: Filesystem>(userdata: *mut c_void) -> sys::FuseLLOps {
    sys::FuseLLOps {
        userdata,
        init: Some(op_init::<F>),
        destroy: Some(op_destroy::<F>),
        lookup: Some(op_lookup::<F>),
        forget: Some(op_forget::<F>),
        getattr: Some(op_getattr::<F>),
        setattr: Some(op_setattr::<F>),
        readlink: Some(op_readlink::<F>),
        mknod: Some(op_mknod::<F>),
        mkdir: Some(op_mkdir::<F>),
        unlink: Some(op_unlink::<F>),
        rmdir: Some(op_rmdir::<F>),
        symlink: Some(op_symlink::<F>),
        rename: Some(op_rename::<F>),
        link: Some(op_link::<F>),
        open: Some(op_open::<F>),
        read: Some(op_read::<F>),
        write: Some(op_write::<F>),
        flush: Some(op_flush::<F>),
        release: Some(op_release::<F>),
        fsync: Some(op_fsync::<F>),
        opendir: Some(op_opendir::<F>),
        readdir: Some(op_readdir::<F>),
        releasedir: Some(op_releasedir::<F>),
        fsyncdir: Some(op_fsyncdir::<F>),
        statfs: Some(op_statfs::<F>),
        setxattr: Some(op_setxattr::<F>),
        getxattr: Some(op_getxattr::<F>),
        listxattr: Some(op_listxattr::<F>),
        removexattr: Some(op_removexattr::<F>),
        access: Some(op_access::<F>),
        create: Some(op_create::<F>),
    }
}

#[cfg(test)]
mod tests {
    use super::{normalize_attr_reply, normalize_statfs_reply};
    use crate::{FUSE_ROOT_ID, S_IFDIR, S_IFREG};
    use fuse_lowlevel_sys::{FuseLLStat, FuseLLStatVfs};

    #[test]
    fn normalizes_root_attr_for_nfs_compatibility() {
        let mut attr = FuseLLStat {
            ino: 0,
            mode: 0o755,
            nlink: 0,
            ..FuseLLStat::default()
        };

        normalize_attr_reply(&mut attr, FUSE_ROOT_ID);

        assert_eq!(attr.ino, FUSE_ROOT_ID);
        assert_eq!(attr.mode, S_IFDIR | 0o755);
        assert_eq!(attr.nlink, 2);
    }

    #[test]
    fn defaults_missing_nlink_for_non_directories() {
        let mut attr = FuseLLStat {
            ino: 42,
            mode: S_IFREG | 0o644,
            nlink: 0,
            ..FuseLLStat::default()
        };
        let ino = attr.ino;

        normalize_attr_reply(&mut attr, ino);

        assert_eq!(attr.ino, 42);
        assert_eq!(attr.mode, S_IFREG | 0o644);
        assert_eq!(attr.nlink, 1);
    }

    #[test]
    fn normalizes_statfs_defaults_for_fuse_t() {
        let mut stat = FuseLLStatVfs::default();

        normalize_statfs_reply(&mut stat);

        assert_eq!(stat.block_size, 4096);
        assert_eq!(stat.name_max, 255);
    }

    #[test]
    fn preserves_explicit_statfs_values() {
        let mut stat = FuseLLStatVfs {
            block_size: 8192,
            name_max: 512,
            ..FuseLLStatVfs::default()
        };

        normalize_statfs_reply(&mut stat);

        assert_eq!(stat.block_size, 8192);
        assert_eq!(stat.name_max, 512);
    }
}
