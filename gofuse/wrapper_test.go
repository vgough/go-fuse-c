package gofuse

import (
	"testing"
)

func TestVersion(t *testing.T) {
	version := FuseVersion()
	if version < 26 {
		t.Errorf("expected version >= 26, got %v", version)
	}
}
