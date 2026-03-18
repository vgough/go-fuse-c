//! Idiomatic Rust wrapper for the FUSE low-level API.
//!
//! The crate keeps the inode-oriented programming model of libfuse while
//! presenting a Rust-first surface:
//!
//! - [`FuseResult`] and [`Errno`] for callback results
//! - [`Filesystem`] trait methods instead of a raw callback table
//! - [`std::ffi::OsStr`] and [`std::ffi::OsString`] for names and link targets
//! - [`std::time::Duration`] and [`std::time::SystemTime`] for time values
//! - [`Session`] and [`MountOptions`] for explicit mount lifecycle control
//!
//! Use [`mount_and_run`] for the common case, or [`Session::mount`] if the
//! caller needs to control the session lifetime directly.

mod error;
mod session;
mod types;

pub use error::{Errno, FuseResult, MountError};
pub use session::{MountOptions, Session, mount_and_run, version};
pub use types::{
    AccessMode, ConnInfo, DirEntryWriter, Entry, FUSE_ROOT_ID, FileInfo, Filesystem, FsFlags,
    Handle, NodeAttr, RequestContext, S_IFDIR, S_IFIFO, S_IFLNK, S_IFREG, SetAttrMask,
    SetAttrRequest, StatFs,
};

#[cfg(target_os = "linux")]
pub mod linux {
    use crate::Filesystem;

    /// Linux-specific extension point for APIs that are not part of the shared
    /// cross-platform trait surface yet.
    pub trait FilesystemExt: Filesystem {}

    impl<T: Filesystem> FilesystemExt for T {}
}

#[cfg(target_os = "macos")]
pub mod macos {
    use crate::Filesystem;

    /// macOS-specific extension point for APIs that are not part of the shared
    /// cross-platform trait surface yet.
    pub trait FilesystemExt: Filesystem {}

    impl<T: Filesystem> FilesystemExt for T {}
}
