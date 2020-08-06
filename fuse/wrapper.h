#ifndef _WRAPPER_H_
#define _WRAPPER_H_

#define _FILE_OFFSET_BITS 64

#if defined(__APPLE__)
// OSXFuse recommends version 26 for new applications.
#define FUSE_USE_VERSION 26
#include <osxfuse/fuse/fuse_lowlevel.h>  // IWYU pragma: export

#else
#define FUSE_USE_VERSION 29
#include <fuse/fuse_lowlevel.h>  // IWYU pragma: export

#endif

#include <sys/types.h>  // for off_t

// Mounts the filesystem and runs the FUSE event loop.
// This call does not return until the filesystem is unmounted.
// Returns an error code, or 0 on success.
//
// Takes ownership of the arguments, using free() to release them.
// int MountAndRun(int id, int argc, char *argv[]);
struct fuse_args *ParseArgs(int argc, char *argv[]);
char *ParseMountpoint(struct fuse_args *args);
struct fuse_chan *Mount(const char *mountpoint, struct fuse_args *args);
struct fuse_session *NewSession(int id, struct fuse_args *args, struct fuse_chan *ch);
int Run(struct fuse_session *se, struct fuse_chan *ch, const char *mountpoint);
void Exit(struct fuse_session *se, struct fuse_chan *ch, const char *mountpoint);

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
void FillTimespec(struct timespec *out, time_t sec, unsigned long nsec);

// enable_bridge_test_mode turns on bridge-level test interceptors.
// This should only be called in test code, and cannot be turned off once enabled.
void enable_bridge_test_mode();

int reply_buf(fuse_req_t req, char *buf, size_t size);

#endif  // _WRAPPER_H_
