package fuse

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var fs RawFileSystem
var fsId int

func TestMain(m *testing.M) {
	fs = NewMemFs()
	fsId = RegisterRawFs(fs)

	r := m.Run()

	DeregisterRawFs(fsId)
	os.Exit(r)
}

func TestVersion(t *testing.T) {
	version := Version()
	Convey("Fuse version", t, func() {
		So(version, ShouldBeGreaterThanOrEqualTo, 26)
	})
}

func TestLookup(t *testing.T) {
	fileEnt, _ := fs.Mknod(1, "exists", 0444, 0)

	Convey("Lookup invalid inode", t, func() {
		bridgeLookup(fsId, 1000, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *replyErr:
				So(r.err, ShouldEqual, ENOENT)

			default:
				t.Errorf("Unexpected reply: %#v", r)
			}
			return int(OK)
		})
	})

	Convey("Lookup invalid file", t, func() {
		bridgeLookup(fsId, 1, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *replyErr:
				So(r.err, ShouldEqual, ENOENT)

			default:
				t.Errorf("Unexpected reply: %#v", r)
			}
			return int(OK)
		})
	})

	Convey("Lookup valid file", t, func() {
		bridgeLookup(fsId, 1, "exists", func(id int, r interface{}) int {
			So(r, ShouldHaveSameTypeAs, &replyEntry{})
			return int(OK)
		})
	})

	Convey("Lookup invalid node type", t, func() {
		// Pass a file inode as the directory.
		bridgeLookup(fsId, fileEnt.Ino, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *replyErr:
				So(r.err, ShouldEqual, ENOTDIR)

			default:
				t.Errorf("Unexpected reply: %#v", r)
			}
			return int(OK)
		})
	})
}

func TestForget(t *testing.T) {
	Convey("Forget uses reply_none", t, func() {
		bridgeForget(fsId, 100, 1, func(id int, r interface{}) int {
			So(r, ShouldHaveSameTypeAs, &replyNone{})
			return int(OK)
		})
	})
}

func TestGetAttr(t *testing.T) {
	Convey("GetAttr on existing directory", t, func() {
		bridgeGetAttr(fsId, 1, func(id int, r interface{}) int {
			So(r, ShouldHaveSameTypeAs, &replyAttr{})
			a := r.(*replyAttr)
			stat := a.attr
			So(stat, ShouldNotBeNil)
			So(stat.st_ino, ShouldEqual, 1)
			return int(OK)
		})
	})

	Convey("GetAttr on nonexistant node", t, func() {
		bridgeGetAttr(fsId, 1000, func(id int, r interface{}) int {
			switch r := r.(type) {
			case *replyErr:
				So(r.err, ShouldEqual, ENOENT)

			default:
				t.Errorf("Unexpected reply: %#v", r)
			}
			return int(OK)
		})
	})
}

func TestStatFs(t *testing.T) {
	Convey("StatFs on undefined inode", t, func() {
		bridgeStatFs(fsId, 0, func(id int, r interface{}) int {
			So(r, ShouldHaveSameTypeAs, &replyStatFs{})
			return int(OK)
		})
	})
}
