go-fuse-c
=========

CGO wrapper for FUSE C low-level API.

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
