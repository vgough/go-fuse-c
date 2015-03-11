#include "wrapper.h"

#include <assert.h>
#include <errno.h>        // IWYU pragma: keep
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

static const int MAGIC_NUM = 0xA239DE71;

struct test_fuse_req {
  int magic;
  int req_id;
  int userdata;
};

fuse_req_t new_fuse_test_req(int id, int userdata) {
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

// Test wrappers which call Go function handlers.
int test_ll_Reply_Err(fuse_req_t req, int err) { return ll_Reply_Err(fuse_test_req_id(req), err); }
void test_ll_Reply_None(fuse_req_t req) { ll_Reply_None(fuse_test_req_id(req)); }
int test_ll_Reply_Entry(fuse_req_t req, const struct fuse_entry_param *e) {
  return ll_Reply_Entry(fuse_test_req_id(req), (struct fuse_entry_param *)e);
}
int test_ll_Reply_Create(fuse_req_t req, const struct fuse_entry_param *e,
                         const struct fuse_file_info *fi) {
  return ll_Reply_Create(fuse_test_req_id(req), (struct fuse_entry_param *)e,
                         (struct fuse_file_info *)fi);
}
int test_ll_Reply_Attr(fuse_req_t req, const struct stat *attr, double attr_timeout) {
  return ll_Reply_Attr(fuse_test_req_id(req), (struct stat *)attr, attr_timeout);
}
int test_ll_Reply_Readlink(fuse_req_t req, const char *link) {
  return ll_Reply_Readlink(fuse_test_req_id(req), (char *)link);
}
int test_ll_Reply_Open(fuse_req_t req, const struct fuse_file_info *fi) {
  return ll_Reply_Open(fuse_test_req_id(req), (struct fuse_file_info *)fi);
}
int test_ll_Reply_Write(fuse_req_t req, size_t count) {
  return ll_Reply_Write(fuse_test_req_id(req), count);
}
int test_ll_Reply_Buf(fuse_req_t req, const char *buf, size_t size) {
  return ll_Reply_Buf(fuse_test_req_id(req), (char *)buf, size);
}
int test_ll_Reply_Statfs(fuse_req_t req, const struct statvfs *stbuf) {
  return ll_Reply_Statfs(fuse_test_req_id(req), (struct statvfs *)stbuf);
}
int test_ll_Reply_Xattr(fuse_req_t req, size_t count) {
  return ll_Reply_Xattr(fuse_test_req_id(req), count);
}
size_t test_ll_Add_Direntry(fuse_req_t req, char *buf, size_t bufsize, const char *name,
                            const struct stat *stbuf, off_t off) {
  return ll_Add_Direntry(fuse_test_req_id(req), buf, bufsize, (char *)name, (struct stat *)stbuf,
                         off);
}
void *test_ll_Req_Userdata(fuse_req_t req) {
  struct test_fuse_req *r = (struct test_fuse_req *)req;
  assert(r->magic == MAGIC_NUM);
  return &r->userdata;
}

// FUSE redirect functions.
// By default, these point to the functions above, which capture the results and make them
// available to the unit tests.
// When configured for production mode, in bridge_init, these functions point to the raw FUSE
// functions.
static int (*ll_reply_err)(fuse_req_t req, int err) = test_ll_Reply_Err;
static void (*ll_reply_none)(fuse_req_t req) = test_ll_Reply_None;
static int (*ll_reply_entry)(fuse_req_t req,
                             const struct fuse_entry_param *e) = test_ll_Reply_Entry;
static int (*ll_reply_create)(fuse_req_t req, const struct fuse_entry_param *e,
                              const struct fuse_file_info *fi) = test_ll_Reply_Create;
static int (*ll_reply_attr)(fuse_req_t req, const struct stat *attr,
                            double attr_timeout) = test_ll_Reply_Attr;
static int (*ll_reply_readlink)(fuse_req_t req, const char *link) = test_ll_Reply_Readlink;
static int (*ll_reply_open)(fuse_req_t req, const struct fuse_file_info *fi) = test_ll_Reply_Open;
static int (*ll_reply_write)(fuse_req_t req, size_t count) = test_ll_Reply_Write;
static int (*ll_reply_buf)(fuse_req_t req, const char *buf, size_t size) = test_ll_Reply_Buf;
static int (*ll_reply_statfs)(fuse_req_t req, const struct statvfs *stbuf) = test_ll_Reply_Statfs;
static int (*ll_reply_xattr)(fuse_req_t req, size_t count) = test_ll_Reply_Xattr;
static size_t (*ll_add_direntry)(fuse_req_t req, char *buf, size_t bufsize, const char *name,
                                 const struct stat *stbuf, off_t off) = test_ll_Add_Direntry;
static void *(*ll_req_userdata)(fuse_req_t req) = test_ll_Req_Userdata;
static const struct fuse_ctx *(*ll_req_ctx)(fuse_req_t req);

// Returns 0 on success.
int DirBufAdd(struct DirBuf *db, const char *name, fuse_ino_t ino, int mode, off_t next) {
  struct stat stbuf = emptyStat;
  stbuf.st_ino = ino;
  stbuf.st_mode = mode;
  stbuf.st_uid = getuid();
  stbuf.st_gid = getgid();

  char *buf = db->buf + db->offset;
  size_t left = db->size - db->offset;
  size_t size = ll_add_direntry(db->req, buf, left, name, &stbuf, next);
  if (size < left) {
    db->offset += size;
    return 0;
  }

  return 1;
}

// The Init call first configures all FUSE wrappers to point to the real FUSE methods.
void bridge_init(void *userdata, struct fuse_conn_info *conn) {
  ll_reply_err = fuse_reply_err;
  ll_reply_none = fuse_reply_none;
  ll_reply_entry = fuse_reply_entry;
  ll_reply_create = fuse_reply_create;
  ll_reply_attr = fuse_reply_attr;
  ll_reply_readlink = fuse_reply_readlink;
  ll_reply_open = fuse_reply_open;
  ll_reply_write = fuse_reply_write;
  ll_reply_buf = fuse_reply_buf;
  ll_reply_statfs = fuse_reply_statfs;
  ll_reply_xattr = fuse_reply_xattr;
  ll_add_direntry = fuse_add_direntry;
  ll_req_userdata = fuse_req_userdata;

  int id = *(int *)userdata;
  ll_Init(id, conn);
}

void bridge_destroy(void *userdata) {
  int id = *(int *)userdata;
  ll_Destroy(id);
}

void bridge_lookup(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Lookup(id, parent, (char *)name, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else if (ll_reply_entry(req, &entry) == -ENOENT) {
    // Request aborted, tell filesystem that reference was dropped.
    ll_Forget(id, entry.ino, 1);
  }
}

void bridge_forget(fuse_req_t req, fuse_ino_t ino, unsigned long nlookup) {
  int id = *(int *)ll_req_userdata(req);
  ll_Forget(id, ino, (int)nlookup);
  ll_reply_none(req);
}

void bridge_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  struct stat attr = emptyStat;
  attr.st_uid = getuid();
  attr.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_GetAttr(id, ino, fi, &attr, &attr_timeout);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_attr(req, &attr, attr_timeout);
  }
}

void bridge_setattr(fuse_req_t req, fuse_ino_t ino, struct stat *attr, int to_set,
                    struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  struct stat out = emptyStat;
  out.st_uid = getuid();
  out.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_SetAttr(id, ino, attr, to_set, fi, &out, &attr_timeout);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_attr(req, &out, attr_timeout);
  }
}

void bridge_readlink(fuse_req_t req, fuse_ino_t ino) {
  int id = *(int *)ll_req_userdata(req);
  int err = 0;
  char *link = ll_ReadLink(id, ino, &err);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_readlink(req, link);
  }
  if (link != NULL) {
    free(link);
  }
}

void bridge_mknod(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, dev_t rdev) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mknod(id, parent, (char *)name, mode, rdev, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_entry(req, &entry);
  }
}

void bridge_mkdir(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mkdir(id, parent, (char *)name, mode, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_entry(req, &entry);
  }
}

void bridge_unlink(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Unlink(id, parent, (char *)name);
  ll_reply_err(req, err);
}

void bridge_rmdir(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Rmdir(id, parent, (char *)name);
  ll_reply_err(req, err);
}

void bridge_symlink(fuse_req_t req, const char *link, fuse_ino_t parent, const char *name) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Symlink(id, (char *)link, parent, (char *)name, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_entry(req, &entry);
  }
}

void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                   const char *newname) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Rename(id, parent, (char *)name, newparent, (char *)newname);
  ll_reply_err(req, err);
}

void bridge_link(fuse_req_t req, fuse_ino_t ino, fuse_ino_t newparent, const char *newname) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Link(id, ino, newparent, (char *)newname, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_entry(req, &entry);
  }
}

void bridge_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Open(id, ino, fi);
  if (err != 0) {
    ll_reply_err(req, err);
  } else if (ll_reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_Release(id, ino, fi);
  }
}

void bridge_read(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                 struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  char *buf = malloc(size);
  if (!buf) {
    ll_reply_err(req, EINTR);
  }
  int n = size;
  int err = ll_Read(id, ino, off, fi, buf, &n);
  if (err != 0) {
    ll_reply_err(req, err);
  }

  ll_reply_buf(req, buf, n);
  free(buf);
}

void bridge_write(fuse_req_t req, fuse_ino_t ino, const char *buf, size_t size, off_t off,
                  struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  size_t written = size;
  int err = ll_Write(id, ino, (char *)buf, &written, off, fi);
  if (err == 0) {
    ll_reply_write(req, written);
  } else {
    ll_reply_err(req, err);
  }
}

void bridge_flush(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Flush(id, ino, fi);
  ll_reply_err(req, err);
}

void bridge_release(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Release(id, ino, fi);
  ll_reply_err(req, err);
}

void bridge_fsync(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_FSync(id, ino, datasync, fi);
  ll_reply_err(req, err);
}

void bridge_opendir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_OpenDir(id, ino, fi);
  if (err != 0) {
    ll_reply_err(req, err);
  } else if (ll_reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_ReleaseDir(id, ino, fi);
  }
}

void bridge_readdir(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                    struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  struct DirBuf db;
  db.req = req;
  db.size = size < 4096 ? 4096 : size;
  db.buf = malloc(db.size);
  db.offset = 0;

  int err = ll_ReadDir(id, ino, size, off, fi, &db);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_buf(req, db.buf, db.offset);
  }

  free(db.buf);
}

void bridge_releasedir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_ReleaseDir(id, ino, fi);
  ll_reply_err(req, err);
}

void bridge_fsyncdir(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_FSyncDir(id, ino, datasync, fi);
  ll_reply_err(req, err);
}

void bridge_statfs(fuse_req_t req, fuse_ino_t ino) {
  int id = *(int *)ll_req_userdata(req);
  struct statvfs stat = emptyStatVfs;
  int err = ll_StatFs(id, ino, &stat);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_statfs(req, &stat);
  }
}

#ifdef __APPLE__
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags, uint32_t position) {
  if (position != 0) {
    ll_reply_err(req, EPERM);
  }
#else
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags) {
#endif

  int id = *(int *)ll_req_userdata(req);
  int err = ll_SetXattr(id, ino, (char *)name, (char *)value, size, flags);
  ll_reply_err(req, err);
}

#ifdef __APPLE__
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size,
                     uint32_t position) {
  if (position != 0) {
    ll_reply_err(req, EPERM);
  }
#else
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size) {
#endif

  int id = *(int *)ll_req_userdata(req);
  char *buf = (size > 0) ? malloc(size) : 0;
  int err = ll_GetXattr(id, ino, (char *)name, buf, &size);
  if (err != 0) {
    ll_reply_err(req, err);
    return;
  }

  if (buf != NULL) {
    ll_reply_buf(req, buf, size);
    free(buf);
  } else {
    ll_reply_xattr(req, size);
  }
}

void bridge_listxattr(fuse_req_t req, fuse_ino_t ino, size_t size) {
  int id = *(int *)ll_req_userdata(req);
  char *buf = (size > 0) ? malloc(size) : 0;
  int err = ll_ListXattr(id, ino, buf, &size);
  if (err != 0) {
    ll_reply_err(req, err);
    return;
  }

  if (buf != NULL) {
    ll_reply_buf(req, buf, size);
    free(buf);
  } else {
    ll_reply_xattr(req, size);
  }
}

void bridge_removexattr(fuse_req_t req, fuse_ino_t ino, const char *name) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_RemoveXattr(id, ino, (char *)name);
  ll_reply_err(req, err);
}

void bridge_access(fuse_req_t req, fuse_ino_t ino, int mask) {
  int id = *(int *)ll_req_userdata(req);
  int err = ll_Access(id, ino, mask);
  ll_reply_err(req, err);
}

void bridge_create(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode,
                   struct fuse_file_info *fi) {
  int id = *(int *)ll_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Create(id, parent, (char *)name, mode, fi, &entry);
  if (err != 0) {
    ll_reply_err(req, err);
  } else {
    ll_reply_create(req, &entry, fi);
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

static struct fuse_lowlevel_ops bridge_ll_ops = {
    .init = bridge_init,
    .destroy = bridge_destroy,
    .lookup = bridge_lookup,
    .forget = bridge_forget,
    .getattr = bridge_getattr,
    .setattr = bridge_setattr,
    .readlink = bridge_readlink,
    .mknod = bridge_mknod,
    .mkdir = bridge_mkdir,
    .unlink = bridge_unlink,
    .rmdir = bridge_rmdir,
    .symlink = bridge_symlink,
    .rename = bridge_rename,
    .link = bridge_link,
    .open = bridge_open,
    .read = bridge_read,
    .write = bridge_write,
    .flush = bridge_flush,
    .release = bridge_release,
    .fsync = bridge_fsync,
    .opendir = bridge_opendir,
    .readdir = bridge_readdir,
    .releasedir = bridge_releasedir,
    .fsyncdir = bridge_fsyncdir,
    .statfs = bridge_statfs,
    .setxattr = bridge_setxattr,
    .getxattr = bridge_getxattr,
    .listxattr = bridge_listxattr,
    .removexattr = bridge_removexattr,
    .access = bridge_access,
    .create = bridge_create,
    //.getlk
    //.setlk
    //.bmap
    //.ioctl
    //.poll
    //.write_buf
    //.retrieve_reply
    //.forget_multi
    //.flock
    //.fallocate
};

const struct fuse_lowlevel_ops *getStandardBridgeOps() { return &bridge_ll_ops; }
