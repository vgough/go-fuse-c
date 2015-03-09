go-fuse-c
=========

CGO wrapper for FUSE C low-level API.

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

