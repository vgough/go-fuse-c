#ifndef FUSE_LOWLEVEL_SYS_WRAPPER_H
#define FUSE_LOWLEVEL_SYS_WRAPPER_H

#define _FILE_OFFSET_BITS 64

#if defined(__APPLE__)
#define FUSE_USE_VERSION 26
#else
#define FUSE_USE_VERSION 35
#endif

#include <fuse_lowlevel.h>

#include <stdint.h>
#include <sys/types.h>

typedef fuse_req_t FuseLLReq;

typedef struct {
  uint64_t secs;
  uint32_t nanos;
} FuseLLDuration;

typedef struct {
  uint64_t ino;
  int64_t size;
  uint32_t mode;
  uint32_t nlink;
  uint32_t uid;
  uint32_t gid;
  int64_t atime_secs;
  uint32_t atime_nanos;
  int64_t ctime_secs;
  uint32_t ctime_nanos;
  int64_t mtime_secs;
  uint32_t mtime_nanos;
} FuseLLStat;

typedef struct {
  uint64_t block_size;
  uint64_t blocks;
  uint64_t blocks_free;
  uint64_t files;
  uint64_t files_free;
  uint64_t fsid;
  uint64_t flags;
  uint64_t name_max;
} FuseLLStatVfs;

typedef struct {
  int flags;
  int writepage;
  uint64_t handle;
  uint64_t lock_owner;
} FuseLLFileInfo;

typedef struct {
  int proto_major;
  int proto_minor;
  unsigned int max_write;
  unsigned int max_readahead;
} FuseLLConnInfo;

typedef struct {
  uint32_t uid;
  uint32_t gid;
  int32_t pid;
  uint32_t umask;
} FuseLLRequestContext;

typedef struct {
  uint64_t ino;
  uint64_t generation;
  FuseLLStat attr;
  FuseLLDuration attr_timeout;
  FuseLLDuration entry_timeout;
} FuseLLEntry;

typedef void (*FuseLLBufFreeFn)(void *data, const unsigned char *buf, size_t len);

typedef struct {
  const unsigned char *data;
  size_t len;
  FuseLLBufFreeFn free_fn;
  void *free_data;
} FuseLLBuf;

typedef struct FuseLLDirBuf {
  fuse_req_t req;
  char *buf;
  size_t size;
  size_t offset;
} FuseLLDirBuf;

typedef struct FuseLLSession FuseLLSession;

typedef struct FuseLLOps {
  void *userdata;

  void (*init)(void *userdata, FuseLLConnInfo *conn);
  void (*destroy)(void *userdata);

  int (*lookup)(void *userdata, FuseLLReq req, uint64_t parent, const char *name, FuseLLEntry *out);
  void (*forget)(void *userdata, FuseLLReq req, uint64_t ino, uint64_t nlookup);
  int (*getattr)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLFileInfo *fi,
                 FuseLLStat *out, FuseLLDuration *timeout);
  int (*setattr)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLStat *attr, int to_set,
                 const FuseLLFileInfo *fi, FuseLLStat *out, FuseLLDuration *timeout);
  int (*readlink)(void *userdata, FuseLLReq req, uint64_t ino, FuseLLBuf *out);
  int (*mknod)(void *userdata, FuseLLReq req, uint64_t parent, const char *name, uint32_t mode,
               uint32_t rdev, FuseLLEntry *out);
  int (*mkdir)(void *userdata, FuseLLReq req, uint64_t parent, const char *name, uint32_t mode,
               FuseLLEntry *out);
  int (*unlink)(void *userdata, FuseLLReq req, uint64_t parent, const char *name);
  int (*rmdir)(void *userdata, FuseLLReq req, uint64_t parent, const char *name);
  int (*symlink)(void *userdata, FuseLLReq req, const char *link, uint64_t parent, const char *name,
                 FuseLLEntry *out);
  int (*rename)(void *userdata, FuseLLReq req, uint64_t parent, const char *name, uint64_t newparent,
                const char *newname, uint32_t flags);
  int (*link)(void *userdata, FuseLLReq req, uint64_t ino, uint64_t newparent, const char *name,
              FuseLLEntry *out);
  int (*open)(void *userdata, FuseLLReq req, uint64_t ino, FuseLLFileInfo *fi);
  int (*read)(void *userdata, FuseLLReq req, uint64_t ino, size_t size, int64_t off,
              const FuseLLFileInfo *fi, FuseLLBuf *out);
  int (*write)(void *userdata, FuseLLReq req, uint64_t ino, const unsigned char *buf, size_t size,
               int64_t off, const FuseLLFileInfo *fi, size_t *written);
  int (*flush)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLFileInfo *fi);
  int (*release)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLFileInfo *fi);
  int (*fsync)(void *userdata, FuseLLReq req, uint64_t ino, int datasync, const FuseLLFileInfo *fi);
  int (*opendir)(void *userdata, FuseLLReq req, uint64_t ino, FuseLLFileInfo *fi);
  int (*readdir)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLFileInfo *fi, int64_t off,
                 size_t size, FuseLLDirBuf *db);
  int (*releasedir)(void *userdata, FuseLLReq req, uint64_t ino, const FuseLLFileInfo *fi);
  int (*fsyncdir)(void *userdata, FuseLLReq req, uint64_t ino, int datasync,
                  const FuseLLFileInfo *fi);
  int (*statfs)(void *userdata, FuseLLReq req, uint64_t ino, FuseLLStatVfs *out);
  int (*setxattr)(void *userdata, FuseLLReq req, uint64_t ino, const char *name,
                  const unsigned char *value, size_t size, int flags);
  int (*getxattr)(void *userdata, FuseLLReq req, uint64_t ino, const char *name, FuseLLBuf *out);
  int (*listxattr)(void *userdata, FuseLLReq req, uint64_t ino, FuseLLBuf *out);
  int (*removexattr)(void *userdata, FuseLLReq req, uint64_t ino, const char *name);
  int (*access)(void *userdata, FuseLLReq req, uint64_t ino, int mask);
  int (*create)(void *userdata, FuseLLReq req, uint64_t parent, const char *name, uint32_t mode,
                FuseLLFileInfo *fi, FuseLLEntry *out);
} FuseLLOps;

int fuse_ll_version(void);
FuseLLSession *fuse_ll_session_mount(int argc, const char *const *argv, const FuseLLOps *ops);
int fuse_ll_session_run(FuseLLSession *session);
void fuse_ll_session_unmount(FuseLLSession *session);
void fuse_ll_session_destroy(FuseLLSession *session);
int fuse_ll_req_ctx(FuseLLReq req, FuseLLRequestContext *out);
int fuse_ll_dirbuf_add(FuseLLDirBuf *db, const char *name, uint64_t ino, uint32_t mode, int64_t next);

#endif  // FUSE_LOWLEVEL_SYS_WRAPPER_H
