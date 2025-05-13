package main

import (
	"log/slog"
	"os"

	"github.com/vgough/go-fuse-c/fuse"
)

func main() {
	// Configure slog to use debug level only if stderr is a terminal
	logLevel := slog.LevelInfo
	if fileInfo, _ := os.Stderr.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		logLevel = slog.LevelDebug
	}
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(logHandler))

	args := os.Args
	// The implementation lives in fuse/memfs, because an in-memory filesystem
	// is also useful in testing.
	ops := fuse.NewMemFS()
	if res := fuse.MountAndRun(args, ops); res < 0 {
		slog.Error("failed", "err", res)
	}
}
