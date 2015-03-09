#ifndef _WRAPPER_H_
#define _WRAPPER_H_

#define FUSE_USE_VERSION 29
#define _FILE_OFFSET_BITS 64

#if defined(__APPLE__)
#include <osxfuse/fuse/fuse_lowlevel.h>  // IWYU pragma: export
#else

#include <fuse/fuse_lowlevel.h>  // IWYU pragma: export
#endif

#include <sys/types.h>  // for off_t

// Mounts the filesystem and runs the FUSE event loop.
// This call does not return until the filesystem is unmounted.
// Returns an error code, or 0 on success.
//
// Takes ownership of the arguments, using free() to release them.
int MountAndRun(int id, int argc, char *argv[], const struct fuse_lowlevel_ops *ops);

// Standard bridge ops, which forward to Go bridge methods.
const struct fuse_lowlevel_ops *getStandardBridgeOps();

struct DirBuf {
  fuse_req_t req;
  char *buf;
  size_t size;

  size_t offset;
};

// Returns 0 on success.
int DirBufAdd(struct DirBuf *db, const char *name, fuse_ino_t ino, int mode, off_t next);

// Helpers to copy time values into timespec.
// This avoids typedef related issues.
void fill_timespec(struct timespec *out, time_t sec, unsigned long nsec);

#endif  // _WRAPPER_H_
