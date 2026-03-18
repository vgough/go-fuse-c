# Rust Port

This workspace contains the in-progress Rust port of `go-fuse-c`.

## Crates

- `fuse-lowlevel-sys`: the unsafe shim layer that adapts `libfuse3` on Linux and `fuse-t` on macOS to a stable Rust-facing ABI.
- `fuse-lowlevel`: the safe public crate that exposes a synchronous low-level `Filesystem` trait, `Result`-based callbacks, `OsStr` names, `Duration` cache timeouts, and `SystemTime` timestamps.

## Go-to-Rust API Mapping

- `FileSystem` interface -> `Filesystem` trait
- `Status` -> `Errno`
- `MountAndRun(args, fs)` -> `mount_and_run(fs, mountpoint, MountOptions)`
- `DefaultFileSystem` -> trait default methods on `Filesystem`
- `Entry` -> `Entry`
- `InoAttr` -> `NodeAttr`
- `StatVFS` -> `StatFs`
- `time.Time` timestamps -> `SystemTime`
- timeout `float64` values -> `Duration`
- `string` entry names -> `&OsStr` and `OsString`

## Examples

- `cargo run -p fuse-lowlevel --example hello -- /tmp/mountpoint`
- `cargo run -p fuse-lowlevel --example hello -- -d -f /tmp/mountpoint`
- `cargo run -p fuse-lowlevel --example memfs -- /tmp/mountpoint`

## Linux Docker Smoke Test

On a Linux host with Docker and `/dev/fuse` available:

- `bash rust/tests/linux/run-in-docker.sh`
- `make rust-docker-test`

The container smoke test builds the Rust `hello` example, mounts it through `libfuse3`, reads the
mounted `hello` file, and then unmounts cleanly.
