use fuse_lowlevel::{
    DirEntryWriter, Entry, Errno, FileInfo, Filesystem, FuseResult, MountOptions, NodeAttr,
    RequestContext, S_IFDIR, S_IFREG, SetAttrMask, StatFs, mount_and_run,
};
use std::collections::HashMap;
use std::ffi::{OsStr, OsString};
use std::os::unix::ffi::OsStrExt;
use std::sync::Mutex;
use std::time::{Duration, SystemTime};

const ATTR_TIMEOUT: Duration = Duration::from_secs(60);

#[derive(Clone)]
struct MemDir {
    parent: Option<u64>,
    nodes: HashMap<OsString, u64>,
}

#[derive(Clone, Default)]
struct MemFile {
    data: Vec<u8>,
}

#[derive(Clone)]
struct INode {
    id: u64,
    dir: Option<MemDir>,
    file: Option<MemFile>,
    ctime: SystemTime,
    mtime: SystemTime,
    mode: u32,
}

impl INode {
    fn stat(&self) -> NodeAttr {
        let mut attr = NodeAttr {
            ino: self.id,
            size: 0,
            mode: self.mode,
            nlink: 1,
            uid: 0,
            gid: 0,
            atime: self.mtime,
            ctime: self.ctime,
            mtime: self.mtime,
            timeout: ATTR_TIMEOUT,
        };
        if let Some(dir) = &self.dir {
            attr.size = dir.nodes.len() as i64;
            attr.mode |= S_IFDIR;
        } else if let Some(file) = &self.file {
            attr.size = file.data.len() as i64;
            attr.mode |= S_IFREG;
        }
        attr
    }
}

struct State {
    generation: u64,
    next_id: u64,
    inodes: HashMap<u64, INode>,
}

struct MemFs {
    state: Mutex<State>,
}

impl MemFs {
    fn new() -> Self {
        let now = SystemTime::now();
        let mut inodes = HashMap::new();
        inodes.insert(
            1,
            INode {
                id: 1,
                dir: Some(MemDir {
                    parent: None,
                    nodes: HashMap::new(),
                }),
                file: None,
                ctime: now,
                mtime: now,
                mode: 0o777,
            },
        );
        Self {
            state: Mutex::new(State {
                generation: now
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_micros() as u64,
                next_id: 2,
                inodes,
            }),
        }
    }

    fn entry(state: &State, node: &INode) -> Entry {
        Entry {
            ino: node.id,
            generation: state.generation,
            attr: node.stat(),
            attr_timeout: ATTR_TIMEOUT,
            entry_timeout: ATTR_TIMEOUT,
        }
    }
}

impl Filesystem for MemFs {
    fn statfs(&self, _req: &RequestContext, _ino: u64) -> FuseResult<StatFs> {
        let state = self.state.lock().expect("memfs state poisoned");
        Ok(StatFs {
            block_size: 4096,
            files: state.inodes.len() as u64,
            name_max: 255,
            ..StatFs::default()
        })
    }

    fn getattr(
        &self,
        _req: &RequestContext,
        ino: u64,
        _fi: Option<&FileInfo>,
    ) -> FuseResult<NodeAttr> {
        let state = self.state.lock().expect("memfs state poisoned");
        state.inodes.get(&ino).map(INode::stat).ok_or(Errno::ENOENT)
    }

    fn setattr(
        &self,
        _req: &RequestContext,
        ino: u64,
        attr: &NodeAttr,
        mask: SetAttrMask,
        _fi: Option<&FileInfo>,
    ) -> FuseResult<NodeAttr> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        let node = state.inodes.get_mut(&ino).ok_or(Errno::ENOENT)?;
        if (mask & SetAttrMask::MODE).0 != 0 {
            node.mode = attr.mode;
        }
        if (mask & SetAttrMask::MTIME).0 != 0 {
            node.mtime = attr.mtime;
        }
        if (mask & SetAttrMask::MTIME_NOW).0 != 0 {
            node.mtime = SystemTime::now();
        }
        if (mask & SetAttrMask::SIZE).0 != 0 {
            let file = node.file.as_mut().ok_or(Errno::EISDIR)?;
            let size = attr.size.max(0) as usize;
            file.data.resize(size, 0);
        }
        Ok(node.stat())
    }

    fn lookup(&self, _req: &RequestContext, parent: u64, name: &OsStr) -> FuseResult<Entry> {
        let state = self.state.lock().expect("memfs state poisoned");
        let parent_node = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
        let dir = parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
        let child = dir.nodes.get(name).ok_or(Errno::ENOENT)?;
        let node = state.inodes.get(child).ok_or(Errno::ENOENT)?;
        Ok(Self::entry(&state, node))
    }

    fn mknod(
        &self,
        _req: &RequestContext,
        parent: u64,
        name: &OsStr,
        mode: u32,
        _rdev: u32,
    ) -> FuseResult<Entry> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        {
            let parent_node = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
            let dir = parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            if dir.nodes.contains_key(name) {
                return Err(Errno::EEXIST);
            }
        }

        let id = state.next_id;
        state.next_id += 1;
        let now = SystemTime::now();
        let node = INode {
            id,
            dir: None,
            file: Some(MemFile::default()),
            ctime: now,
            mtime: now,
            mode,
        };
        let entry = Self::entry(&state, &node);
        state.inodes.insert(id, node);
        let parent_node = state.inodes.get_mut(&parent).ok_or(Errno::ENOENT)?;
        parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .insert(name.to_os_string(), id);
        Ok(entry)
    }

    fn create(
        &self,
        req: &RequestContext,
        parent: u64,
        name: &OsStr,
        mode: u32,
        _fi: &mut FileInfo,
    ) -> FuseResult<Entry> {
        self.mknod(req, parent, name, mode, 0)
    }

    fn mkdir(
        &self,
        _req: &RequestContext,
        parent: u64,
        name: &OsStr,
        mode: u32,
    ) -> FuseResult<Entry> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        {
            let parent_node = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
            let dir = parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            if dir.nodes.contains_key(name) {
                return Err(Errno::EEXIST);
            }
        }

        let id = state.next_id;
        state.next_id += 1;
        let now = SystemTime::now();
        let node = INode {
            id,
            dir: Some(MemDir {
                parent: Some(parent),
                nodes: HashMap::new(),
            }),
            file: None,
            ctime: now,
            mtime: now,
            mode,
        };
        let entry = Self::entry(&state, &node);
        state.inodes.insert(id, node);
        let parent_node = state.inodes.get_mut(&parent).ok_or(Errno::ENOENT)?;
        parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .insert(name.to_os_string(), id);
        Ok(entry)
    }

    fn open(&self, _req: &RequestContext, ino: u64, _fi: &mut FileInfo) -> FuseResult<()> {
        let state = self.state.lock().expect("memfs state poisoned");
        let node = state.inodes.get(&ino).ok_or(Errno::ENOENT)?;
        if node.file.is_none() {
            return Err(Errno::EISDIR);
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
        let state = self.state.lock().expect("memfs state poisoned");
        let node = state.inodes.get(&ino).ok_or(Errno::ENOENT)?;
        let file = node.file.as_ref().ok_or(Errno::EISDIR)?;
        let off = off.max(0) as usize;
        if off >= file.data.len() {
            return Ok(Vec::new());
        }
        let end = (off + size).min(file.data.len());
        Ok(file.data[off..end].to_vec())
    }

    fn write(
        &self,
        _req: &RequestContext,
        data: &[u8],
        ino: u64,
        off: i64,
        _fi: &FileInfo,
    ) -> FuseResult<usize> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        let node = state.inodes.get_mut(&ino).ok_or(Errno::ENOENT)?;
        let file = node.file.as_mut().ok_or(Errno::EISDIR)?;
        let off = off.max(0) as usize;
        let required = off + data.len();
        if required > file.data.len() {
            file.data.resize(required, 0);
        }
        file.data[off..off + data.len()].copy_from_slice(data);
        Ok(data.len())
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
        let state = self.state.lock().expect("memfs state poisoned");
        let node = state.inodes.get(&ino).ok_or(Errno::ENOENT)?;
        let dir = node.dir.as_ref().ok_or(Errno::ENOTDIR)?;

        let mut index = 1i64;
        if index > off && !writer.add(OsStr::new("."), ino, node.mode | S_IFDIR, index) {
            return Ok(());
        }
        index += 1;
        if let Some(parent) = dir.parent {
            if index > off && !writer.add(OsStr::new(".."), parent, S_IFDIR, index) {
                return Ok(());
            }
            index += 1;
        }

        for (name, child) in &dir.nodes {
            if index > off {
                let child_node = state.inodes.get(child).ok_or(Errno::ENOENT)?;
                if !writer.add(name, *child, child_node.mode, index) {
                    return Ok(());
                }
            }
            index += 1;
        }

        Ok(())
    }

    fn unlink(&self, _req: &RequestContext, parent: u64, name: &OsStr) -> FuseResult<()> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        let child_id = {
            let parent_node = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
            let dir = parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            *dir.nodes.get(name).ok_or(Errno::ENOENT)?
        };
        let child = state.inodes.get(&child_id).ok_or(Errno::ENOENT)?;
        if child.file.is_none() {
            return Err(Errno::EISDIR);
        }
        state.inodes.remove(&child_id);
        let parent_node = state.inodes.get_mut(&parent).ok_or(Errno::ENOENT)?;
        parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .remove(name);
        Ok(())
    }

    fn rmdir(&self, _req: &RequestContext, parent: u64, name: &OsStr) -> FuseResult<()> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        let child_id = {
            let parent_node = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
            let dir = parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            *dir.nodes.get(name).ok_or(Errno::ENOENT)?
        };
        let child = state.inodes.get(&child_id).ok_or(Errno::ENOENT)?;
        let dir = child.dir.as_ref().ok_or(Errno::ENOTDIR)?;
        if !dir.nodes.is_empty() {
            return Err(Errno::ENOTEMPTY);
        }
        state.inodes.remove(&child_id);
        let parent_node = state.inodes.get_mut(&parent).ok_or(Errno::ENOENT)?;
        parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .remove(name);
        Ok(())
    }

    fn rename(
        &self,
        _req: &RequestContext,
        parent: u64,
        name: &OsStr,
        newparent: u64,
        newname: &OsStr,
        _flags: u32,
    ) -> FuseResult<()> {
        let mut state = self.state.lock().expect("memfs state poisoned");
        let node_id = {
            let old_parent = state.inodes.get(&parent).ok_or(Errno::ENOENT)?;
            let dir = old_parent.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            *dir.nodes.get(name).ok_or(Errno::ENOENT)?
        };
        let replaced = {
            let new_parent_node = state.inodes.get(&newparent).ok_or(Errno::ENOENT)?;
            let dir = new_parent_node.dir.as_ref().ok_or(Errno::ENOTDIR)?;
            dir.nodes.get(newname).copied()
        };
        if let Some(existing) = replaced {
            let existing_node = state.inodes.get(&existing).ok_or(Errno::ENOENT)?;
            if existing_node.file.is_none() {
                return Err(Errno::EISDIR);
            }
            state.inodes.remove(&existing);
        }
        let old_parent_node = state.inodes.get_mut(&parent).ok_or(Errno::ENOENT)?;
        old_parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .remove(name);
        let new_parent_node = state.inodes.get_mut(&newparent).ok_or(Errno::ENOENT)?;
        new_parent_node
            .dir
            .as_mut()
            .ok_or(Errno::ENOTDIR)?
            .nodes
            .insert(newname.to_os_string(), node_id);
        Ok(())
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
    let program = args.next().unwrap_or_else(|| OsString::from("memfs"));
    let (mountpoint, raw_args) = match split_mountpoint(args.collect()) {
        Some(values) => values,
        None => {
            eprintln!("usage: memfs [fuse args...] <mountpoint>");
            std::process::exit(2);
        }
    };

    let mut options = MountOptions::new().program_name(program);
    for arg in raw_args {
        options = options.raw_arg(arg);
    }

    if let Err(err) = mount_and_run(MemFs::new(), &mountpoint, options) {
        eprintln!("failed to run memfs: {err}");
        std::process::exit(1);
    }
}
