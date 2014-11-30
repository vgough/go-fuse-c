#ifndef _WRAPPER_H_
#define _WRAPPER_H_

#define FUSE_USE_VERSION 29
#define _FILE_OFFSET_BITS 64

#include <fuse/fuse_lowlevel.h>  // IWYU pragma: export
#include <sys/types.h>           // for off_t

int MountAndRun(int id, int argc, char *argv[]);

struct DirBuf {
  fuse_req_t req;
  char *buf;
  size_t size;

  size_t offset;
};

// Returns 0 on success.
int DirBufAdd(struct DirBuf *db, const char *name, fuse_ino_t ino, int mode,
              off_t next);

#endif  // _WRAPPER_H_
