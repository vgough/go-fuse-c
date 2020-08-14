#include "wrapper.h"

#include <assert.h>
#include <errno.h>        // IWYU pragma: keep
#include <stdbool.h>      // for bool
#include <stdint.h>       // for uint64_t
#include <stdlib.h>       // for free
#include <sys/stat.h>     // for stat, mode_t, dev_t
#include <sys/statvfs.h>  // for statvfs
#include <unistd.h>       // for off_t, getgid, getuid

#include "_cgo_export.h"  // IWYU pragma: keep

// Bridge methods are FUSE C callbacks.  The callbacks pass through to the
// corresponding Go function and then translate the return value into the
// appropriate FUSE result call.

// TODO: log error result from all fuse_reply_* methods.

static const struct stat emptyStat;
static const struct fuse_entry_param emptyEntry;
static const struct statvfs emptyStatVfs;

// bridge_test_mode is set to TRUE when testing the bridge interface.
static bool bridge_test_mode = false;

void enable_bridge_test_mode() { bridge_test_mode = true; }

static const int MAGIC_NUM = 0xA239DE71;

struct test_fuse_req {
  int    magic;
  int    req_id;
  char * userdata;
};

fuse_req_t new_fuse_test_req(int id, char *userdata) {
  struct test_fuse_req *req = malloc(sizeof(struct test_fuse_req));
  req->magic = MAGIC_NUM;
  req->req_id = id;
  req->userdata = userdata;
  return (fuse_req_t)req;
}

void free_fuse_test_req(fuse_req_t req) {
  struct test_fuse_req *r = (struct test_fuse_req *)req;
  assert(r->magic == MAGIC_NUM);
  free(r);
}

int fuse_test_req_id(fuse_req_t req) {
  struct test_fuse_req *r = (struct test_fuse_req *)req;
  assert(r->magic == MAGIC_NUM);
  return r->req_id;
}

char * get_fsid(fuse_req_t req) {
  if (bridge_test_mode) {
    struct test_fuse_req *r = (struct test_fuse_req *)req;
    assert(r->magic == MAGIC_NUM);
    return r->userdata;
  }

  return (char *)fuse_req_userdata(req);
}

static int reply_err(fuse_req_t req, int err) {
  if (bridge_test_mode) {
    return test_Reply_Err(fuse_test_req_id(req), err);
  }

  return fuse_reply_err(req, err);
}

static int reply_entry(fuse_req_t req, struct fuse_entry_param *ent) {
  if (bridge_test_mode) {
    return test_Reply_Entry(fuse_test_req_id(req), ent);
  }

  return fuse_reply_entry(req, ent);
}

static void reply_none(fuse_req_t req) {
  if (bridge_test_mode) {
    test_Reply_None(fuse_test_req_id(req));
    return;
  }

  fuse_reply_none(req);
  return;
}

static int reply_create(fuse_req_t req, struct fuse_entry_param *ent, struct fuse_file_info *fi) {
  if (bridge_test_mode) {
    return test_Reply_Create(fuse_test_req_id(req), ent, fi);
  }
  return fuse_reply_create(req, ent, fi);
}

static int reply_attr(fuse_req_t req, struct stat *attr, double timeout) {
  if (bridge_test_mode) {
    return test_Reply_Attr(fuse_test_req_id(req), attr, timeout);
  }
  return fuse_reply_attr(req, attr, timeout);
}

static int reply_readlink(fuse_req_t req, char *link) {
  if (bridge_test_mode) {
    return test_Reply_Readlink(fuse_test_req_id(req), link);
  }
  return fuse_reply_readlink(req, link);
}

static int reply_open(fuse_req_t req, struct fuse_file_info *fi) {
  if (bridge_test_mode) {
    return test_Reply_Open(fuse_test_req_id(req), fi);
  }

  return fuse_reply_open(req, fi);
}

static int reply_write(fuse_req_t req, size_t count) {
  if (bridge_test_mode) {
    return test_Reply_Write(fuse_test_req_id(req), count);
  }

  return fuse_reply_write(req, count);
}

int reply_buf(fuse_req_t req, char *buf, size_t size) {
  if (bridge_test_mode) {
    return test_Reply_Buf(fuse_test_req_id(req), buf, size);
  }

  return fuse_reply_buf(req, buf, size);
}

static int reply_statfs(fuse_req_t req, struct statvfs *stbuf) {
  if (bridge_test_mode) {
    return test_Reply_Statfs(fuse_test_req_id(req), stbuf);
  }

  return fuse_reply_statfs(req, stbuf);
}

static int reply_xattr(fuse_req_t req, size_t count) {
  if (bridge_test_mode) {
    return test_Reply_Xattr(fuse_test_req_id(req), count);
  }

  return fuse_reply_xattr(req, count);
}

// The Init call first configures all FUSE wrappers to point to the real FUSE methods.
void bridge_init(void *userdata, struct fuse_conn_info *conn) {
  ll_Init((char *) userdata, conn);
}

void bridge_destroy(void *userdata) {
  ll_Destroy((char *) userdata);
}

void bridge_lookup(fuse_req_t req, fuse_ino_t parent, const char *name) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Lookup(mp, parent, (char *)name, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else if (reply_entry(req, &entry) == -ENOENT) {
    // Request aborted, tell filesystem that reference was dropped.
    ll_Forget(mp, entry.ino, 1);
  }
}

void bridge_forget(fuse_req_t req, fuse_ino_t ino, unsigned long nlookup) {
  char * mp = get_fsid(req);
  ll_Forget(mp, ino, (int)nlookup);
  reply_none(req);
}

void bridge_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  struct stat attr = emptyStat;
  attr.st_uid = getuid();
  attr.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_GetAttr(mp, ino, fi, &attr, &attr_timeout);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_attr(req, &attr, attr_timeout);
  }
}

void bridge_setattr(fuse_req_t req, fuse_ino_t ino, struct stat *attr, int to_set,
                    struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  struct stat out = emptyStat;
  out.st_uid = getuid();
  out.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_SetAttr(mp, ino, attr, to_set, fi, &out, &attr_timeout);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_attr(req, &out, attr_timeout);
  }
}

void bridge_readlink(fuse_req_t req, fuse_ino_t ino) {
  char * mp = get_fsid(req);
  int err = 0;
  char *link = ll_ReadLink(mp, ino, &err);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_readlink(req, link);
  }
  if (link != NULL) {
    free(link);
  }
}

void bridge_mknod(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, dev_t rdev) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mknod(mp, parent, (char *)name, mode, rdev, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_entry(req, &entry);
  }
}

void bridge_mkdir(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mkdir(mp, parent, (char *)name, mode, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_entry(req, &entry);
  }
}

void bridge_unlink(fuse_req_t req, fuse_ino_t parent, const char *name) {
  char * mp = get_fsid(req);
  int err = ll_Unlink(mp, parent, (char *)name);
  reply_err(req, err);
}

void bridge_rmdir(fuse_req_t req, fuse_ino_t parent, const char *name) {
  char * mp = get_fsid(req);
  int err = ll_Rmdir(mp, parent, (char *)name);
  reply_err(req, err);
}

void bridge_symlink(fuse_req_t req, const char *link, fuse_ino_t parent, const char *name) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Symlink(mp, (char *)link, parent, (char *)name, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_entry(req, &entry);
  }
}

void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                   const char *newname) {
  char * mp = get_fsid(req);
  int err = ll_Rename(mp, parent, (char *)name, newparent, (char *)newname);
  reply_err(req, err);
}

void bridge_link(fuse_req_t req, fuse_ino_t ino, fuse_ino_t newparent, const char *newname) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Link(mp, ino, newparent, (char *)newname, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_entry(req, &entry);
  }
}

void bridge_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_Open(mp, ino, fi);
  if (err != 0) {
    reply_err(req, err);
  } else if (reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_Release(mp, ino, fi);
  }
}

void bridge_read(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                 struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_Read(mp, req, ino, size, off, fi);
  if (err != 0) {
    reply_err(req, err);
  }
}

void bridge_write(fuse_req_t req, fuse_ino_t ino, const char *buf, size_t size, off_t off,
                  struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  size_t written = size;
  int err = ll_Write(mp, ino, (char *)buf, &written, off, fi);
  if (err == 0) {
    reply_write(req, written);
  } else {
    reply_err(req, err);
  }
}

void bridge_flush(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_Flush(mp, ino, fi);
  reply_err(req, err);
}

void bridge_release(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_Release(mp, ino, fi);
  reply_err(req, err);
}

void bridge_fsync(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_FSync(mp, ino, datasync, fi);
  reply_err(req, err);
}

void bridge_opendir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_OpenDir(mp, ino, fi);
  if (err != 0) {
    reply_err(req, err);
  } else if (reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_ReleaseDir(mp, ino, fi);
  }
}

void bridge_readdir(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                    struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  struct DirBuf db;
  db.req = req;
  db.size = size < 4096 ? 4096 : size;
  db.buf = malloc(db.size);
  db.offset = 0;

  int err = ll_ReadDir(mp, ino, size, off, fi, &db);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_buf(req, db.buf, db.offset);
  }

  free(db.buf);
}

void bridge_releasedir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_ReleaseDir(mp, ino, fi);
  reply_err(req, err);
}

void bridge_fsyncdir(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  int err = ll_FSyncDir(mp, ino, datasync, fi);
  reply_err(req, err);
}

void bridge_statfs(fuse_req_t req, fuse_ino_t ino) {
  char * mp = get_fsid(req);
  struct statvfs stat = emptyStatVfs;
  int err = ll_StatFS(mp, ino, &stat);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_statfs(req, &stat);
  }
}

#ifdef __APPLE__
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags, uint32_t position) {
  if (position != 0) {
    reply_err(req, EPERM);
  }
#else
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags) {
#endif

  char * mp = get_fsid(req);
  int err = ll_SetXAttr(mp, ino, (char *)name, (char *)value, size, flags);
  reply_err(req, err);
}

#ifdef __APPLE__
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size,
                     uint32_t position) {
  if (position != 0) {
    reply_err(req, EPERM);
  }
#else
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size) {
#endif

  char * mp = get_fsid(req);
  char *buf = (size > 0) ? malloc(size) : 0;
  int err = ll_GetXAttr(mp, ino, (char *)name, buf, &size);
  if (err != 0) {
    reply_err(req, err);
    return;
  }

  if (buf != NULL) {
    reply_buf(req, buf, size);
    free(buf);
  } else {
    reply_xattr(req, size);
  }
}

void bridge_listxattr(fuse_req_t req, fuse_ino_t ino, size_t size) {
  char * mp = get_fsid(req);
  char *buf = (size > 0) ? malloc(size) : 0;
  int err = ll_ListXAttr(mp, ino, buf, &size);
  if (err != 0) {
    reply_err(req, err);
    return;
  }

  if (buf != NULL) {
    reply_buf(req, buf, size);
    free(buf);
  } else {
    reply_xattr(req, size);
  }
}

void bridge_removexattr(fuse_req_t req, fuse_ino_t ino, const char *name) {
  char * mp = get_fsid(req);
  int err = ll_RemoveXAttr(mp, ino, (char *)name);
  reply_err(req, err);
}

void bridge_access(fuse_req_t req, fuse_ino_t ino, int mask) {
  char * mp = get_fsid(req);
  int err = ll_Access(mp, ino, mask);
  reply_err(req, err);
}

void bridge_create(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode,
                   struct fuse_file_info *fi) {
  char * mp = get_fsid(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Create(mp, parent, (char *)name, mode, fi, &entry);
  if (err != 0) {
    reply_err(req, err);
  } else {
    reply_create(req, &entry, fi);
  }
}

void bridge_getlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock);
void bridge_setlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock,
                  int sleep);
void bridge_bmap(fuse_req_t req, fuse_ino_t ino, size_t blocksize, uint64_t idx);
void bridge_ioctl(fuse_req_t req, fuse_ino_t ino, int cmd, void *arg, struct fuse_file_info *fi,
                  unsigned flags, const void *in_buf, size_t in_bufsz, size_t out_bufsz);

#if 0  // Not available on OSX.  Make conditional upon version & platform?
void bridge_poll(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi,
                 struct fuse_pollhandle *ph);
void bridge_write_buf(fuse_req_t req, fuse_ino_t ino, struct fuse_bufvec *bufv, off_t off,
                      struct fuse_file_info *fi);
void bridge_retrieve_reply(fuse_req_t req, void *cookie, fuse_ino_t ino, off_t offset,
                           struct fuse_bufvec *bufv);
void bridge_forget_multi(fuse_req_t req, size_t count, struct fuse_forget_data *forgets);
#endif

void bridge_flock(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, int op);
void bridge_fallocate(fuse_req_t req, fuse_ino_t ino, int mode, off_t offset, off_t length,
                      struct fuse_file_info *fi);
