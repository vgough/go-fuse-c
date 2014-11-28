package main

import (
	"fmt"
	"github.com/vgough/go-fuse-c/gofuse"
	"os"
)

type HelloFs struct {
}

func (h *HelloFs) Init(*gofuse.FuseConnInfo) {
}

func (h *HelloFs) Destroy() {
}

func (h *HelloFs) Lookup(parent int64, name string) (
  err gofuse.Status, entry *gofuse.FuseEntryParam) {

  if parent != 1 || name != "hello" {
    return gofuse.ENOENT, nil
  }

  e := &gofuse.FuseEntryParam{
    Ino: 2,
    AttrTimeout: 1.0,
    EntryTimeout: 1.0,
  }
  // TODO: fill out stat

  return gofuse.OK, e
}

func main() {
	args := os.Args
	fmt.Println(args)
	ops := &HelloFs{}
	fmt.Println("fuse main returned", gofuse.MountAndRun(args, ops))
}
