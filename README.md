go-fuse-c
=========

[![Build Status](https://gitlab.com/vgough/go-fuse-c/badges/master/pipeline.svg)](https://gitlab.com/vgough/go-fuse-c/pipelines)
[![GoDoc](https://godoc.org/github.com/vgough/go-fuse-c?status.svg)](http://godoc.org/github.com/vgough/go-fuse-c/fuse)

CGO wrapper for FUSE C low-level API.

# Purpose

After running into trouble getting a couple "pure-Go" FUSE wrappers working, I
decided that the most practical approach was to wrap the C library.  This would
allow reuse of system-specific libraries (like OSXFuse) and avoid ongoing work
of porting fixes from multiple client libraries.  If your system has a C-API
compatible libfuse, then it is likely that this will work with it.

Additionally, I want access to the FUSE Low-Level API, which deals with inodes,
rather than the Path based API.  Although the Path based API makes it easy to
write simple filesystems, the Low Level API is more powerful and makes it easier
to make the filesystem behave like a built-in Posix filesystem.

## Alternatives

For more FUSE-related references, see [Resources Related to FUSE](https://github.com/koding/awesome-fuse-fs)

* [GoFuse](https://github.com/hanwen/go-fuse): The GoFuse library is a little
difficult to use, even if you're familiar with FUSE and Go.

* [Bazil Fuse](https://github.com/bazil/fuse): Bazil has a low-level API which
is similar to this library, however it is lacks good examples and has been
mostly dead for years.  4 years ago (April 2015), I spent time figuring out all
the interfaces used by Bazil in order to make an in-memory FS example.  That PR
request is still pending [add in-memory example](https://github.com/bazil/fuse/pull/83).

* [jacobsa/fuse](https://github.com/jacobsa/fuse): Has lots of samples. Don't
 have any experience with it myself.

# Examples

To try the hello example, which corresponds to hello low-level API example
that comes with FUSE:

````
go build example/hello/hello.go
./hello /tmp/mountpoint
````

Also provided is a slightly more interesting in-memory filesystem:

````
go build example/memfs/memfs.go
./memfs /tmp/mountpoint
````

# Development Notes

## Bridge functions and filesystem handles

The C bridge functions handle the initial FUSE operation callbacks.  They call
through to static Go functions which are exported in the bridge code.  These
static functions lookup the filesystem from the provided filesystem handle and
pass control to the filesystem implementation.

Integer filesystem handles are used instead of pointers as it is bad form to
hold pointers to Go structures in C.

## Testing

Bridge methods are normally called from FUSE.  However the bridge methods are
also available in Go for testing purposes.  Since types such as `fuse_req` are
opaque, the FUSE reply methods cannot be called during tests.  An internal
function `enable_bridge_test_mode` is used to switch to using test reply methods
so that the C / Go interfaces can be tested without running FUSE code.

Run `go test -v ./...` to execute tests.
