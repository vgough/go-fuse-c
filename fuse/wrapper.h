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

// Initialize fuse_args 
struct fuse_args *InitArgs(int argc, char *argv[]);

// Returns mountpoint from fuse arguments
char *ParseMountpoint(struct fuse_args *args);

// Mount file system and returns fuse_channel
struct fuse_chan *Mount(char *mountpoint, struct fuse_args *args);

// Creates a fuse session with provided mount point and arguments
struct fuse_session *NewSession(char *mountpoint, struct fuse_args *args, struct fuse_chan *ch);

// Runs the FUSE event loop.
// This call does not return until the filesystem is unmounted.
// Returns an error code, or 0 on success.
int Run(char *mountpoint, struct fuse_session *se, struct fuse_chan *ch);

// Sets exit flag of the session and unmounts the filesystem
void Exit(char *mountpoint, struct fuse_session *se, struct fuse_chan *ch);

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
