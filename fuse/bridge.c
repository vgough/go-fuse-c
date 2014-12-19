#include "wrapper.h"

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

void bridge_init(void *userdata, struct fuse_conn_info *conn) {
  int id = *(int *)userdata;
  ll_Init(id, conn);
}

void bridge_destroy(void *userdata) {
  int id = *(int *)userdata;
  ll_Destroy(id);
}

void bridge_lookup(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Lookup(id, parent, (char *)name, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else if (fuse_reply_entry(req, &entry) == -ENOENT) {
    // Request aborted, tell filesystem that reference was dropped.
    ll_Forget(id, entry.ino, 1);
  }
}

void bridge_forget(fuse_req_t req, fuse_ino_t ino, unsigned long nlookup) {
  int id = *(int *)fuse_req_userdata(req);
  ll_Forget(id, ino, (int)nlookup);
  fuse_reply_none(req);
}

void bridge_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  struct stat attr = emptyStat;
  attr.st_uid = getuid();
  attr.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_GetAttr(id, ino, fi, &attr, &attr_timeout);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_attr(req, &attr, attr_timeout);
  }
}

void bridge_setattr(fuse_req_t req, fuse_ino_t ino, struct stat *attr, int to_set,
                    struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  struct stat out = emptyStat;
  out.st_uid = getuid();
  out.st_gid = getgid();
  double attr_timeout = 1.0;
  int err = ll_SetAttr(id, ino, attr, to_set, fi, &out, &attr_timeout);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_attr(req, &out, attr_timeout);
  }
}

void bridge_readlink(fuse_req_t req, fuse_ino_t ino) {
  int id = *(int *)fuse_req_userdata(req);
  int err = 0;
  char *link = ll_ReadLink(id, ino, &err);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_readlink(req, link);
  }
  if (link != NULL) {
    free(link);
  }
}

void bridge_mknod(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, dev_t rdev) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mknod(id, parent, (char *)name, mode, rdev, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_entry(req, &entry);
  }
}

void bridge_mkdir(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Mkdir(id, parent, (char *)name, mode, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_entry(req, &entry);
  }
}

void bridge_unlink(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Unlink(id, parent, (char *)name);
  fuse_reply_err(req, err);
}

void bridge_rmdir(fuse_req_t req, fuse_ino_t parent, const char *name) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Rmdir(id, parent, (char *)name);
  fuse_reply_err(req, err);
}

void bridge_symlink(fuse_req_t req, const char *link, fuse_ino_t parent, const char *name) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Symlink(id, (char *)link, parent, (char *)name, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_entry(req, &entry);
  }
}

void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                   const char *newname) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Rename(id, parent, (char *)name, newparent, (char *)newname);
  fuse_reply_err(req, err);
}

void bridge_link(fuse_req_t req, fuse_ino_t ino, fuse_ino_t newparent, const char *newname) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Link(id, ino, newparent, (char *)newname, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_entry(req, &entry);
  }
}

void bridge_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Open(id, ino, fi);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else if (fuse_reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_Release(id, ino, fi);
  }
}

void bridge_read(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                 struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  char *buf = malloc(size);
  if (!buf) {
    fuse_reply_err(req, EINTR);
  }
  int n = size;
  int err = ll_Read(id, ino, off, fi, buf, &n);
  if (err != 0) {
    fuse_reply_err(req, err);
  }

  fuse_reply_buf(req, buf, n);
  free(buf);
}

void bridge_write(fuse_req_t req, fuse_ino_t ino, const char *buf, size_t size, off_t off,
                  struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  size_t written = size;
  int err = ll_Write(id, ino, (char *)buf, &written, off, fi);
  if (err == 0) {
    fuse_reply_write(req, written);
  } else {
    fuse_reply_err(req, err);
  }
}

void bridge_flush(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Flush(id, ino, fi);
  fuse_reply_err(req, err);
}

void bridge_release(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Release(id, ino, fi);
  fuse_reply_err(req, err);
}

void bridge_fsync(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_FSync(id, ino, datasync, fi);
  fuse_reply_err(req, err);
}

void bridge_opendir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_OpenDir(id, ino, fi);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else if (fuse_reply_open(req, fi) == -ENOENT) {
    // Request aborted, let Go wrapper know.
    ll_ReleaseDir(id, ino, fi);
  }
}

void bridge_readdir(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                    struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  struct DirBuf db;
  db.req = req;
  db.size = size < 4096 ? 4096 : size;
  db.buf = malloc(db.size);
  db.offset = 0;

  int err = ll_ReadDir(id, ino, size, off, fi, &db);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_buf(req, db.buf, db.offset);
  }

  free(db.buf);
}

void bridge_releasedir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_ReleaseDir(id, ino, fi);
  fuse_reply_err(req, err);
}

void bridge_fsyncdir(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_FSyncDir(id, ino, datasync, fi);
  fuse_reply_err(req, err);
}

void bridge_statfs(fuse_req_t req, fuse_ino_t ino) {
  int id = *(int *)fuse_req_userdata(req);
  struct statvfs stat = emptyStatVfs;
  int err = ll_StatFs(id, ino, &stat);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_statfs(req, &stat);
  }
}

void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags);
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size);
void bridge_listxattr(fuse_req_t req, fuse_ino_t ino, size_t size);
void bridge_removexattr(fuse_req_t req, fuse_ino_t ino, const char *name);

void bridge_access(fuse_req_t req, fuse_ino_t ino, int mask) {
  int id = *(int *)fuse_req_userdata(req);
  int err = ll_Access(id, ino, mask);
  fuse_reply_err(req, err);
}

void bridge_create(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode,
                   struct fuse_file_info *fi) {
  int id = *(int *)fuse_req_userdata(req);
  struct fuse_entry_param entry = emptyEntry;
  entry.attr.st_uid = getuid();
  entry.attr.st_gid = getgid();
  int err = ll_Create(id, parent, (char *)name, mode, fi, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
  } else {
    fuse_reply_create(req, &entry, fi);
  }
}

void bridge_getlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock);
void bridge_setlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock,
                  int sleep);
void bridge_bmap(fuse_req_t req, fuse_ino_t ino, size_t blocksize, uint64_t idx);
void bridge_ioctl(fuse_req_t req, fuse_ino_t ino, int cmd, void *arg, struct fuse_file_info *fi,
                  unsigned flags, const void *in_buf, size_t in_bufsz, size_t out_bufsz);
void bridge_poll(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi,
                 struct fuse_pollhandle *ph);
void bridge_write_buf(fuse_req_t req, fuse_ino_t ino, struct fuse_bufvec *bufv, off_t off,
                      struct fuse_file_info *fi);
void bridge_retrieve_reply(fuse_req_t req, void *cookie, fuse_ino_t ino, off_t offset,
                           struct fuse_bufvec *bufv);
void bridge_forget_multi(fuse_req_t req, size_t count, struct fuse_forget_data *forgets);
void bridge_flock(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, int op);
void bridge_fallocate(fuse_req_t req, fuse_ino_t ino, int mode, off_t offset, off_t length,
                      struct fuse_file_info *fi);

struct fuse_lowlevel_ops bridge_ll_ops = {.init = bridge_init,
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
                                          //.setxattr
                                          //.getxattr
                                          //.listxattr
                                          //.removexattr
                                          //.access
                                          //.create
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
