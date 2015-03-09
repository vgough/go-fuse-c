package fuse

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVersion(t *testing.T) {
	version := Version()
	Convey("Fuse version", t, func() {
		So(version, ShouldBeGreaterThanOrEqualTo, 26)
	})
}

func TestLookup(t *testing.T) {
	fs := NewMemFs()
	fsId := RegisterRawFs(fs)
	defer DeregisterRawFs(fsId)
	fileEnt, _ := fs.Mknod(1, "exists", 0444, 0)

	Convey("Lookup invalid inode", t, func() {
		BridgeLookup(fsId, 1000, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *ReplyErr:
				So(r.err, ShouldEqual, ENOENT)

			default:
				t.Errorf("Unexpected reply: %v", r)
			}
			return int(OK)
		})
	})

	Convey("Lookup invalid file", t, func() {
		BridgeLookup(fsId, 1, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *ReplyErr:
				So(r.err, ShouldEqual, ENOENT)

			default:
				t.Errorf("Unexpected reply: %v", r)
			}
			return int(OK)
		})
	})

	Convey("Lookup valid file", t, func() {
		BridgeLookup(fsId, 1, "exists", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *ReplyErr:
				So(r.err, ShouldEqual, OK)

			default:
				t.Errorf("Unexpected reply: %v", r)
			}
			return int(OK)
		})
	})

	Convey("Lookup invalid node type", t, func() {
		// Pass a file inode as the directory.
		BridgeLookup(fsId, fileEnt.Ino, "test", func(id int, r interface{}) int {
			switch r := r.(type) {
			case *ReplyErr:
				So(r.err, ShouldEqual, ENOTDIR)

			default:
				t.Errorf("Unexpected reply: %v", r)
			}
			return int(OK)
		})
	})
}

func TestForget(t *testing.T) {
	fs := NewMemFs()
	fsId := RegisterRawFs(fs)
	defer DeregisterRawFs(fsId)

	Convey("Forget uses reply_none", t, func() {
		BridgeForget(fsId, 100, 1, func(id int, r interface{}) int {
			So(r, ShouldHaveSameTypeAs, &ReplyNone{})
			return int(OK)
		})
	})
}
