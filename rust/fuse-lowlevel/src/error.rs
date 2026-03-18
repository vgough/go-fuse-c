use std::error::Error;
use std::fmt;

/// Error number returned from a filesystem callback.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct Errno(i32);

/// Result type used by filesystem operations in this crate.
pub type FuseResult<T> = std::result::Result<T, Errno>;

impl Errno {
    pub const OK: Self = Self(0);
    pub const EACCES: Self = Self(13);
    pub const EBUSY: Self = Self(16);
    pub const EEXIST: Self = Self(17);
    pub const EINVAL: Self = Self(22);
    pub const EIO: Self = Self(5);
    pub const ENOENT: Self = Self(2);
    pub const ENOTDIR: Self = Self(20);
    pub const ENOTEMPTY: Self = Self(39);
    pub const EPERM: Self = Self(1);
    pub const ERANGE: Self = Self(34);
    pub const EXDEV: Self = Self(18);
    pub const EBADF: Self = Self(9);
    pub const ENODEV: Self = Self(19);
    pub const EROFS: Self = Self(30);
    pub const EISDIR: Self = Self(21);
    pub const ENOMEM: Self = Self(12);

    #[cfg(target_os = "macos")]
    pub const ENOSYS: Self = Self(78);
    #[cfg(target_os = "linux")]
    pub const ENOSYS: Self = Self(38);

    #[cfg(target_os = "macos")]
    pub const ENODATA: Self = Self(93);
    #[cfg(target_os = "linux")]
    pub const ENODATA: Self = Self(61);

    /// Creates an [`Errno`] from a raw platform error number.
    pub const fn from_raw(raw: i32) -> Self {
        Self(raw)
    }

    /// Returns the raw platform error number.
    pub const fn raw(self) -> i32 {
        self.0
    }
}

impl fmt::Display for Errno {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "errno {}", self.0)
    }
}

impl Error for Errno {}

/// Error returned by mount/session management APIs.
///
/// This is intentionally separate from [`Errno`], which is reserved for errors
/// returned from filesystem callbacks back to the kernel.
#[derive(Debug)]
pub enum MountError {
    /// A mount argument contained an interior NUL byte.
    NulByte,
    /// The native FUSE session could not be created or mounted.
    MountFailed,
    /// The FUSE event loop exited with a non-zero status code.
    RunFailed(i32),
}

impl fmt::Display for MountError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::NulByte => f.write_str("mount arguments contain an interior NUL byte"),
            Self::MountFailed => f.write_str("failed to mount FUSE session"),
            Self::RunFailed(code) => write!(f, "FUSE session exited with status {code}"),
        }
    }
}

impl Error for MountError {}
