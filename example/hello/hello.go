package main

import (
	"fmt"
	"github.com/vgough/go-fuse-c/fuse"
	"os"
)

type HelloFs struct {
}

func (h *HelloFs) Init(*fuse.FuseConnInfo) {
}

func (h *HelloFs) Destroy() {
}

func (h *HelloFs) Lookup(parent int64, name string) (
  err fuse.Status, entry *fuse.FuseEntryParam) {

  if parent != 1 || name != "hello" {
    return fuse.ENOENT, nil
  }

  e := &fuse.FuseEntryParam{
    Ino: 2,
    AttrTimeout: 1.0,
    EntryTimeout: 1.0,
  }
  // TODO: fill out stat

  return fuse.OK, e
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", fuse.MountAndRun(args, ops))
}
