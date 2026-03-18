use crate::{Errno, FuseResult};
use std::ffi::{OsStr, OsString};
use std::ops::{BitAnd, BitAndAssign, BitOr, BitOrAssign};
use std::time::{Duration, SystemTime};

/// Root inode number reserved by FUSE.
pub const FUSE_ROOT_ID: u64 = 1;
/// FIFO file type bit.
pub const S_IFIFO: u32 = 0o010000;
/// Directory file type bit.
pub const S_IFDIR: u32 = 0o040000;
/// Regular file type bit.
pub const S_IFREG: u32 = 0o100000;
/// Symbolic-link file type bit.
pub const S_IFLNK: u32 = 0o120000;

/// Opaque file or directory handle stored in `fuse_file_info::fh`.
#[derive(Clone, Copy, Debug, Default, Eq, PartialEq)]
pub struct Handle(pub u64);

/// Access mode portion of an open request.
#[derive(Clone, Copy, Debug, Default, Eq, PartialEq)]
pub struct AccessMode(pub i32);

impl AccessMode {
    /// Read-only access.
    pub const O_RDONLY: Self = Self(0);
    /// Write-only access.
    pub const O_WRONLY: Self = Self(1);
    /// Read/write access.
    pub const O_RDWR: Self = Self(2);
}

/// Filesystem flags returned from [`StatFs`].
#[derive(Clone, Copy, Debug, Default, Eq, PartialEq)]
pub struct FsFlags(pub u64);

impl FsFlags {
    /// Read-only filesystem.
    pub const ST_RDONLY: Self = Self(1);
    /// Ignore setuid/setgid bits.
    pub const ST_NOSUID: Self = Self(2);
}

impl BitOr for FsFlags {
    type Output = Self;
    fn bitor(self, rhs: Self) -> Self::Output {
        Self(self.0 | rhs.0)
    }
}

impl BitOrAssign for FsFlags {
    fn bitor_assign(&mut self, rhs: Self) {
        self.0 |= rhs.0;
    }
}

/// Bitmask describing which fields in a setattr request are valid.
#[derive(Clone, Copy, Debug, Default, Eq, PartialEq)]
pub struct SetAttrMask(pub u32);

impl SetAttrMask {
    pub const MODE: Self = Self(1 << 0);
    pub const UID: Self = Self(1 << 1);
    pub const GID: Self = Self(1 << 2);
    pub const SIZE: Self = Self(1 << 3);
    pub const ATIME: Self = Self(1 << 4);
    pub const MTIME: Self = Self(1 << 5);
    pub const ATIME_NOW: Self = Self(1 << 7);
    pub const MTIME_NOW: Self = Self(1 << 8);
}

impl BitOr for SetAttrMask {
    type Output = Self;
    fn bitor(self, rhs: Self) -> Self::Output {
        Self(self.0 | rhs.0)
    }
}

impl BitOrAssign for SetAttrMask {
    fn bitor_assign(&mut self, rhs: Self) {
        self.0 |= rhs.0;
    }
}

impl BitAnd for SetAttrMask {
    type Output = Self;
    fn bitand(self, rhs: Self) -> Self::Output {
        Self(self.0 & rhs.0)
    }
}

impl BitAndAssign for SetAttrMask {
    fn bitand_assign(&mut self, rhs: Self) {
        self.0 &= rhs.0;
    }
}

/// Process metadata attached to a request by the kernel.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct RequestContext {
    pub uid: u32,
    pub gid: u32,
    pub pid: i32,
    pub umask: u32,
}

/// File-open state exchanged with the kernel.
#[derive(Clone, Debug)]
pub struct FileInfo {
    pub flags: i32,
    pub writepage: bool,
    pub handle: Handle,
    pub lock_owner: u64,
}

impl FileInfo {
    /// Returns only the access mode portion of [`FileInfo::flags`].
    pub fn access_mode(&self) -> AccessMode {
        AccessMode(self.flags & 0x3)
    }
}

/// Connection metadata negotiated with the FUSE client library.
#[derive(Clone, Debug)]
pub struct ConnInfo {
    pub proto_major: i32,
    pub proto_minor: i32,
    pub max_write: u32,
    pub max_readahead: u32,
}

/// Attributes returned for an inode.
#[derive(Clone, Debug)]
pub struct NodeAttr {
    pub ino: u64,
    pub size: i64,
    pub mode: u32,
    pub nlink: u32,
    pub uid: u32,
    pub gid: u32,
    pub atime: SystemTime,
    pub ctime: SystemTime,
    pub mtime: SystemTime,
    pub timeout: Duration,
}

/// Directory entry returned from lookup-style operations.
#[derive(Clone, Debug)]
pub struct Entry {
    pub ino: u64,
    pub generation: u64,
    pub attr: NodeAttr,
    pub attr_timeout: Duration,
    pub entry_timeout: Duration,
}

/// Filesystem statistics returned from `statfs`.
#[derive(Clone, Debug, Default)]
pub struct StatFs {
    pub block_size: u64,
    pub blocks: u64,
    pub blocks_free: u64,
    pub files: u64,
    pub files_free: u64,
    pub fsid: u64,
    pub flags: FsFlags,
    pub name_max: u64,
}

/// Structured setattr request payload.
#[derive(Clone, Debug)]
pub struct SetAttrRequest {
    pub attr: NodeAttr,
    pub mask: SetAttrMask,
}

/// Writer used by [`Filesystem::readdir`] to append directory entries.
pub trait DirEntryWriter {
    /// Adds a single directory entry to the current reply buffer.
    ///
    /// Returns `false` when there is not enough space left in the current
    /// kernel reply buffer.
    fn add(&mut self, name: &OsStr, ino: u64, mode: u32, next: i64) -> bool;
}

/// Low-level FUSE filesystem implementation.
///
/// The trait follows the libfuse low-level API closely while using Rust types,
/// borrowed path/name values, and [`FuseResult`] for errors. Unsupported
/// operations default to [`Errno::ENOSYS`] unless the callback is explicitly
/// defined as a no-op.
pub trait Filesystem: Send + Sync + 'static {
    /// Initializes the filesystem before request handling begins.
    fn init(&self, _conn: &mut ConnInfo) {}

    /// Cleans up the filesystem after the session exits.
    fn destroy(&self) {}

    /// Returns filesystem statistics for the given inode.
    fn statfs(&self, _req: &RequestContext, _ino: u64) -> FuseResult<StatFs> {
        Err(Errno::ENOSYS)
    }

    /// Resolves `name` inside `parent` and returns its directory entry.
    fn lookup(&self, _req: &RequestContext, _parent: u64, _name: &OsStr) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Drops kernel references to an inode.
    fn forget(&self, _req: &RequestContext, _ino: u64, _nlookup: u64) {}

    /// Releases an open file handle.
    fn release(&self, _req: &RequestContext, _ino: u64, _fi: &FileInfo) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Releases an open directory handle.
    fn releasedir(&self, _req: &RequestContext, _ino: u64, _fi: &FileInfo) -> FuseResult<()> {
        Ok(())
    }

    /// Flushes file data and/or metadata to stable storage.
    fn fsync(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _data_only: bool,
        _fi: &FileInfo,
    ) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Flushes directory data and/or metadata to stable storage.
    fn fsyncdir(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _data_only: bool,
        _fi: &FileInfo,
    ) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Handles per-close flush requests for a file.
    fn flush(&self, _req: &RequestContext, _ino: u64, _fi: &FileInfo) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Returns current attributes for an inode.
    fn getattr(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _fi: Option<&FileInfo>,
    ) -> FuseResult<NodeAttr> {
        Err(Errno::ENOSYS)
    }

    /// Applies a partial attribute update to an inode.
    fn setattr(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _attr: &NodeAttr,
        _mask: SetAttrMask,
        _fi: Option<&FileInfo>,
    ) -> FuseResult<NodeAttr> {
        Err(Errno::ENOSYS)
    }

    /// Reads the target path of a symbolic link.
    fn readlink(&self, _req: &RequestContext, _ino: u64) -> FuseResult<OsString> {
        Err(Errno::ENOSYS)
    }

    /// Emits directory entries starting at the given offset.
    fn readdir(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _fi: &FileInfo,
        _off: i64,
        _size: usize,
        _writer: &mut dyn DirEntryWriter,
    ) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Creates a filesystem node under `parent`.
    fn mknod(
        &self,
        _req: &RequestContext,
        _parent: u64,
        _name: &OsStr,
        _mode: u32,
        _rdev: u32,
    ) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Checks access permissions for an inode.
    fn access(&self, _req: &RequestContext, _ino: u64, _mask: i32) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Creates and opens a file in one operation.
    fn create(
        &self,
        _req: &RequestContext,
        _parent: u64,
        _name: &OsStr,
        _mode: u32,
        _fi: &mut FileInfo,
    ) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Opens a file and may install a new [`Handle`].
    fn open(&self, _req: &RequestContext, _ino: u64, _fi: &mut FileInfo) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Opens a directory and may install a new [`Handle`].
    fn opendir(&self, _req: &RequestContext, _ino: u64, _fi: &mut FileInfo) -> FuseResult<()> {
        Ok(())
    }

    /// Reads file data at `off`.
    fn read(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _size: usize,
        _off: i64,
        _fi: &FileInfo,
    ) -> FuseResult<Vec<u8>> {
        Err(Errno::ENOSYS)
    }

    /// Writes file data at `off`.
    fn write(
        &self,
        _req: &RequestContext,
        _data: &[u8],
        _ino: u64,
        _off: i64,
        _fi: &FileInfo,
    ) -> FuseResult<usize> {
        Err(Errno::ENOSYS)
    }

    /// Creates a directory under `parent`.
    fn mkdir(
        &self,
        _req: &RequestContext,
        _parent: u64,
        _name: &OsStr,
        _mode: u32,
    ) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Removes an empty directory.
    fn rmdir(&self, _req: &RequestContext, _parent: u64, _name: &OsStr) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Creates a symbolic link.
    fn symlink(
        &self,
        _req: &RequestContext,
        _link: &OsStr,
        _parent: u64,
        _name: &OsStr,
    ) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Creates a hard link to an existing inode.
    fn link(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _newparent: u64,
        _name: &OsStr,
    ) -> FuseResult<Entry> {
        Err(Errno::ENOSYS)
    }

    /// Renames or moves an entry.
    fn rename(
        &self,
        _req: &RequestContext,
        _parent: u64,
        _name: &OsStr,
        _newparent: u64,
        _newname: &OsStr,
        _flags: u32,
    ) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Removes a non-directory entry.
    fn unlink(&self, _req: &RequestContext, _parent: u64, _name: &OsStr) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Lists extended attribute names for an inode.
    fn listxattrs(&self, _req: &RequestContext, _ino: u64) -> FuseResult<Vec<OsString>> {
        Err(Errno::ENOSYS)
    }

    /// Reads an extended attribute value.
    fn getxattr(&self, _req: &RequestContext, _ino: u64, _name: &OsStr) -> FuseResult<Vec<u8>> {
        Err(Errno::ENOSYS)
    }

    /// Writes an extended attribute value.
    fn setxattr(
        &self,
        _req: &RequestContext,
        _ino: u64,
        _name: &OsStr,
        _value: &[u8],
        _flags: i32,
    ) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }

    /// Removes an extended attribute.
    fn removexattr(&self, _req: &RequestContext, _ino: u64, _name: &OsStr) -> FuseResult<()> {
        Err(Errno::ENOSYS)
    }
}
