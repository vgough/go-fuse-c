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

	Convey("Test Lookup", t, func() {

		Convey("Invalid inode", func() {
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

		Convey("Invalid file", func() {
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
	})
}
