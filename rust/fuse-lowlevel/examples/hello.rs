use fuse_lowlevel::{
    AccessMode, DirEntryWriter, Entry, Errno, FileInfo, Filesystem, FuseResult, MountOptions,
    NodeAttr, RequestContext, S_IFDIR, S_IFREG, StatFs, mount_and_run,
};
use std::ffi::{OsStr, OsString};
use std::os::unix::ffi::OsStrExt;
use std::time::{Duration, SystemTime};

const HELLO_STR: &[u8] = b"Hello World!\n";

struct HelloFs {
    mount_time: SystemTime,
}

impl HelloFs {
    fn stat(&self, ino: u64) -> Option<NodeAttr> {
        let mut attr = NodeAttr {
            ino,
            size: 0,
            mode: 0,
            nlink: 0,
            uid: 0,
            gid: 0,
            atime: self.mount_time,
            ctime: self.mount_time,
            mtime: self.mount_time,
            timeout: Duration::from_secs(1),
        };

        match ino {
            1 => {
                attr.mode = S_IFDIR | 0o755;
                attr.nlink = 2;
            }
            2 => {
                attr.mode = S_IFREG | 0o444;
                attr.nlink = 1;
                attr.size = HELLO_STR.len() as i64;
            }
            _ => return None,
        }

        Some(attr)
    }
}

impl Filesystem for HelloFs {
    fn getattr(
        &self,
        _req: &RequestContext,
        ino: u64,
        _fi: Option<&FileInfo>,
    ) -> FuseResult<NodeAttr> {
        self.stat(ino).ok_or(Errno::ENOENT)
    }

    fn lookup(&self, _req: &RequestContext, parent: u64, name: &OsStr) -> FuseResult<Entry> {
        if parent != 1 || name != OsStr::new("hello") {
            return Err(Errno::ENOENT);
        }

        Ok(Entry {
            ino: 2,
            generation: 1,
            attr: self.stat(2).expect("hello file attr"),
            attr_timeout: Duration::from_secs(1),
            entry_timeout: Duration::from_secs(1),
        })
    }

    fn statfs(&self, _req: &RequestContext, _ino: u64) -> FuseResult<StatFs> {
        Ok(StatFs {
            block_size: 4096,
            files: 2,
            files_free: 0,
            flags: fuse_lowlevel::FsFlags::ST_RDONLY,
            name_max: 255,
            ..StatFs::default()
        })
    }

    fn readdir(
        &self,
        _req: &RequestContext,
        ino: u64,
        _fi: &FileInfo,
        off: i64,
        _size: usize,
        writer: &mut dyn DirEntryWriter,
    ) -> FuseResult<()> {
        if ino != 1 {
            return Err(Errno::ENOTDIR);
        }
        if off < 1 && !writer.add(OsStr::new("."), 1, S_IFDIR, 1) {
            return Ok(());
        }
        if off < 2 && !writer.add(OsStr::new(".."), 1, S_IFDIR, 2) {
            return Ok(());
        }
        if off < 3 {
            let _ = writer.add(OsStr::new("hello"), 2, S_IFREG, 3);
        }
        Ok(())
    }

    fn open(&self, _req: &RequestContext, ino: u64, fi: &mut FileInfo) -> FuseResult<()> {
        if ino != 2 {
            return Err(Errno::EISDIR);
        }
        if fi.access_mode() != AccessMode::O_RDONLY {
            return Err(Errno::EACCES);
        }
        Ok(())
    }

    fn read(
        &self,
        _req: &RequestContext,
        ino: u64,
        size: usize,
        off: i64,
        _fi: &FileInfo,
    ) -> FuseResult<Vec<u8>> {
        if ino != 2 {
            return Err(Errno::ENOENT);
        }
        let off = off.max(0) as usize;
        if off >= HELLO_STR.len() {
            return Ok(Vec::new());
        }
        let end = (off + size).min(HELLO_STR.len());
        Ok(HELLO_STR[off..end].to_vec())
    }
}

fn split_mountpoint(args: Vec<OsString>) -> Option<(OsString, Vec<OsString>)> {
    let mut positional = Vec::new();
    let mut option_takes_value = false;

    for (idx, arg) in args.iter().enumerate() {
        if option_takes_value {
            option_takes_value = false;
            continue;
        }

        let arg = arg.as_os_str();
        if arg == OsStr::new("-o") || arg == OsStr::new("--option") {
            option_takes_value = true;
            continue;
        }

        if !matches!(arg.as_bytes().first(), Some(b'-')) {
            positional.push(idx);
        }
    }

    let mount_idx = match positional.as_slice() {
        [] => return None,
        [idx] => *idx,
        _ if positional[0] == 0 => positional[0],
        _ => *positional.last().expect("positional args are non-empty"),
    };

    let mountpoint = args[mount_idx].clone();
    let raw_args = args
        .into_iter()
        .enumerate()
        .filter_map(|(idx, arg)| (idx != mount_idx).then_some(arg))
        .collect();
    Some((mountpoint, raw_args))
}

fn main() {
    let mut args = std::env::args_os();
    let program = args.next().unwrap_or_else(|| OsString::from("hello"));
    let (mountpoint, raw_args) = match split_mountpoint(args.collect()) {
        Some(values) => values,
        None => {
            eprintln!("usage: hello [fuse args...] <mountpoint>");
            std::process::exit(2);
        }
    };

    let mut options = MountOptions::new().program_name(program);
    for arg in raw_args {
        options = options.raw_arg(arg);
    }

    let fs = HelloFs {
        mount_time: SystemTime::now(),
    };
    if let Err(err) = mount_and_run(fs, &mountpoint, options) {
        eprintln!("failed to run hello filesystem: {err}");
        std::process::exit(1);
    }
}
