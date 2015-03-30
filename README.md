go-fuse-c
=========

[![Build Status](https://travis-ci.org/vgough/go-fuse-c.svg)](https://travis-ci.org/vgough/go-fuse-c)
[![GoDoc](https://godoc.org/github.com/vgough/go-fuse-c?status.svg)](http://godoc.org/github.com/vgough/go-fuse-c/fuse)

CGO wrapper for FUSE C low-level API.

# Purpose

After running into trouble getting a "pure-Go" FUSE wrapper working, I decided that the most
practical approach was to wrap the C library.  This would allow reuse of system-specific libraries
(like OSXFuse) and avoid ongoing work of porting fixes from multiple client libraries.

Additionally, I want access to the FUSE Low-Level API, which deals with inodes, rather than the
Path based API.  Although the Path based API makes it easy to write simple filesystems, the Low
Level API is more powerful and makes it easier to make the filesystem behave like a built-in Posix
filesystem.

# STATUS

2015-03-15: I've recently discovered another "pure-Go" wrapper,
[bazillion/fuse](https://github.com/bazillion/fuse) which wraps the Low-level API and has support
for Linux and OSX.  If this works well, then I may drop go-fuse-c entirely.

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

The C bridge functions handle the initial FUSE operation callbacks.  They call through to static Go
functions which are exported in the bridge code.  These static functions lookup the filesystem from
the provided filesystem handle and pass control to the filesystem implementation.

Integer filesystem handles are used instead of pointers as it is bad form to hold pointers to Go
structures in C.

## Testing

Bridge methods are normally called from FUSE.  However the bridge methods are also available in Go
for testing purposes.  Since types such as `fuse_req` are opaque, the FUSE reply methods cannot be
called during tests.  Function pointers are provided for all FUSE reply operations, which point
to the real FUSE reply operations after the filesystem `init` method is called.  When running unit
tests, the functions point to an implementation which captures the response for validation.

Tests are being written using [GoConvey](https://github.com/smartystreets/goconvey).  Run
`go test -v ./...` to execute tests once.  For an automatically-updating web UI, install and
run `$GOPATH/bin/goconvey`.

