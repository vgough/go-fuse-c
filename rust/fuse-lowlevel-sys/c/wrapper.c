#if defined(__APPLE__)
#define _DARWIN_C_SOURCE
#else
#define _POSIX_C_SOURCE 200809L
#endif

#include "wrapper.h"

#include <errno.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/stat.h>
#include <sys/statvfs.h>
#include <unistd.h>

#if defined(__APPLE__)
#include <fuse.h>
#endif

struct FuseLLSession {
  struct fuse_args args;
  struct fuse_session *session;
  struct fuse_chan *channel;
  char *mountpoint;
  FuseLLOps ops;
  int mounted;
  int signal_handlers;
#if !defined(__APPLE__)
  struct fuse_cmdline_opts cmdline;
#endif
};

static const struct stat EMPTY_STAT;
static const struct fuse_entry_param EMPTY_ENTRY;
static const struct statvfs EMPTY_STATVFS;

static FuseLLOps *session_ops(FuseLLSession *session) { return &session->ops; }

static void buf_release(FuseLLBuf *buf) {
  if (buf == NULL) {
    return;
  }
  if (buf->free_fn != NULL && buf->data != NULL) {
    buf->free_fn(buf->free_data, buf->data, buf->len);
  }
  buf->data = NULL;
  buf->len = 0;
  buf->free_fn = NULL;
  buf->free_data = NULL;
}

static void fill_timespec(struct timespec *out, FuseLLDuration in) {
  out->tv_sec = (time_t)in.secs;
  out->tv_nsec = (long)in.nanos;
}

static void stat_to_native(const FuseLLStat *in, struct stat *out) {
  memset(out, 0, sizeof(*out));
  out->st_ino = (ino_t)in->ino;
  out->st_size = (off_t)in->size;
  out->st_blksize = 4096;
  out->st_mode = (mode_t)in->mode;
  out->st_nlink = (nlink_t)in->nlink;
  out->st_uid = (uid_t)in->uid;
  out->st_gid = (gid_t)in->gid;
#if defined(__APPLE__)
  out->st_atimespec.tv_sec = (time_t)in->atime_secs;
  out->st_atimespec.tv_nsec = (long)in->atime_nanos;
  out->st_ctimespec.tv_sec = (time_t)in->ctime_secs;
  out->st_ctimespec.tv_nsec = (long)in->ctime_nanos;
  out->st_mtimespec.tv_sec = (time_t)in->mtime_secs;
  out->st_mtimespec.tv_nsec = (long)in->mtime_nanos;
#else
  out->st_atim.tv_sec = (time_t)in->atime_secs;
  out->st_atim.tv_nsec = (long)in->atime_nanos;
  out->st_ctim.tv_sec = (time_t)in->ctime_secs;
  out->st_ctim.tv_nsec = (long)in->ctime_nanos;
  out->st_mtim.tv_sec = (time_t)in->mtime_secs;
  out->st_mtim.tv_nsec = (long)in->mtime_nanos;
#endif
}

static void stat_from_native(const struct stat *in, FuseLLStat *out) {
  memset(out, 0, sizeof(*out));
  out->ino = (uint64_t)in->st_ino;
  out->size = (int64_t)in->st_size;
  out->mode = (uint32_t)in->st_mode;
  out->nlink = (uint32_t)in->st_nlink;
  out->uid = (uint32_t)in->st_uid;
  out->gid = (uint32_t)in->st_gid;
#if defined(__APPLE__)
  out->atime_secs = (int64_t)in->st_atimespec.tv_sec;
  out->atime_nanos = (uint32_t)in->st_atimespec.tv_nsec;
  out->ctime_secs = (int64_t)in->st_ctimespec.tv_sec;
  out->ctime_nanos = (uint32_t)in->st_ctimespec.tv_nsec;
  out->mtime_secs = (int64_t)in->st_mtimespec.tv_sec;
  out->mtime_nanos = (uint32_t)in->st_mtimespec.tv_nsec;
#else
  out->atime_secs = (int64_t)in->st_atim.tv_sec;
  out->atime_nanos = (uint32_t)in->st_atim.tv_nsec;
  out->ctime_secs = (int64_t)in->st_ctim.tv_sec;
  out->ctime_nanos = (uint32_t)in->st_ctim.tv_nsec;
  out->mtime_secs = (int64_t)in->st_mtim.tv_sec;
  out->mtime_nanos = (uint32_t)in->st_mtim.tv_nsec;
#endif
}

static void file_info_from_native(const struct fuse_file_info *in, FuseLLFileInfo *out) {
  memset(out, 0, sizeof(*out));
  if (in == NULL) {
    return;
  }
  out->flags = in->flags;
  out->writepage = in->writepage;
  out->handle = in->fh;
  out->lock_owner = in->lock_owner;
}

static void file_info_to_native(const FuseLLFileInfo *in, struct fuse_file_info *out) {
  if (in == NULL || out == NULL) {
    return;
  }
  out->fh = in->handle;
}

static void conn_from_native(const struct fuse_conn_info *in, FuseLLConnInfo *out) {
  out->proto_major = in->proto_major;
  out->proto_minor = in->proto_minor;
  out->max_write = in->max_write;
  out->max_readahead = in->max_readahead;
}

static void conn_to_native(const FuseLLConnInfo *in, struct fuse_conn_info *out) {
  out->max_write = in->max_write;
  out->max_readahead = in->max_readahead;
}

static void entry_to_native(const FuseLLEntry *in, struct fuse_entry_param *out) {
  memset(out, 0, sizeof(*out));
  out->ino = (fuse_ino_t)in->ino;
  out->generation = (unsigned long)(in->generation == 0 ? 1 : in->generation);
  stat_to_native(&in->attr, &out->attr);
  out->attr_timeout = (double)in->attr_timeout.secs + ((double)in->attr_timeout.nanos / 1000000000.0);
  out->entry_timeout =
      (double)in->entry_timeout.secs + ((double)in->entry_timeout.nanos / 1000000000.0);
}

static void statvfs_to_native(const FuseLLStatVfs *in, struct statvfs *out) {
  unsigned long block_size = (unsigned long)(in->block_size == 0 ? 4096 : in->block_size);
  unsigned long name_max = (unsigned long)(in->name_max == 0 ? 255 : in->name_max);

  memset(out, 0, sizeof(*out));
  out->f_bsize = block_size;
  out->f_frsize = block_size;
  out->f_blocks = in->blocks;
  out->f_bfree = in->blocks_free;
  out->f_bavail = in->blocks_free;
  out->f_files = in->files;
  out->f_ffree = in->files_free;
  out->f_favail = in->files_free;
  out->f_fsid = (unsigned long)in->fsid;
  out->f_flag = (unsigned long)in->flags;
  out->f_namemax = name_max;
}

int fuse_ll_version(void) { return fuse_version(); }

int fuse_ll_req_ctx(FuseLLReq req, FuseLLRequestContext *out) {
  const struct fuse_ctx *ctx = fuse_req_ctx(req);
  if (ctx == NULL || out == NULL) {
    return -1;
  }
  out->uid = ctx->uid;
  out->gid = ctx->gid;
  out->pid = ctx->pid;
  out->umask = ctx->umask;
  return 0;
}

int fuse_ll_dirbuf_add(FuseLLDirBuf *db, const char *name, uint64_t ino, uint32_t mode, int64_t next) {
  struct stat stbuf = EMPTY_STAT;
  stbuf.st_ino = (ino_t)ino;
  stbuf.st_mode = (mode_t)mode;
  stbuf.st_uid = getuid();
  stbuf.st_gid = getgid();

  char *buf = db->buf + db->offset;
  size_t left = db->size - db->offset;
  size_t written = fuse_add_direntry(db->req, buf, left, name, &stbuf, (off_t)next);
  if (written < left) {
    db->offset += written;
    return 0;
  }

  return 1;
}

static void bridge_init(void *userdata, struct fuse_conn_info *conn) {
  FuseLLSession *session = (FuseLLSession *)userdata;
  FuseLLOps *ops = session_ops(session);
  if (ops->init == NULL) {
    return;
  }
  FuseLLConnInfo info;
  conn_from_native(conn, &info);
  ops->init(ops->userdata, &info);
  conn_to_native(&info, conn);
}

static void bridge_destroy(void *userdata) {
  FuseLLSession *session = (FuseLLSession *)userdata;
  FuseLLOps *ops = session_ops(session);
  if (ops->destroy != NULL) {
    ops->destroy(ops->userdata);
  }
}

static void bridge_lookup(fuse_req_t req, fuse_ino_t parent, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->lookup == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLEntry entry = {0};
  int err = ops->lookup(ops->userdata, req, parent, name, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  if (fuse_reply_entry(req, &native_entry) == -ENOENT && ops->forget != NULL) {
    ops->forget(ops->userdata, req, entry.ino, 1);
  }
}

static void bridge_forget(fuse_req_t req, fuse_ino_t ino, unsigned long nlookup) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->forget != NULL) {
    ops->forget(ops->userdata, req, ino, nlookup);
  }
  fuse_reply_none(req);
}

static void bridge_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->getattr == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  FuseLLStat stat = {0};
  FuseLLDuration timeout = {0};
  int err = ops->getattr(ops->userdata, req, ino, fi == NULL ? NULL : &info, &stat, &timeout);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct stat native_stat = EMPTY_STAT;
  stat_to_native(&stat, &native_stat);
  double native_timeout = (double)timeout.secs + ((double)timeout.nanos / 1000000000.0);
  fuse_reply_attr(req, &native_stat, native_timeout);
}

static void bridge_setattr(
    fuse_req_t req, fuse_ino_t ino, struct stat *attr, int to_set, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->setattr == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  FuseLLStat input = {0};
  stat_from_native(attr, &input);
  FuseLLStat output = {0};
  FuseLLDuration timeout = {0};
  int err = ops->setattr(ops->userdata, req, ino, &input, to_set, fi == NULL ? NULL : &info, &output,
                         &timeout);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct stat native_stat = EMPTY_STAT;
  stat_to_native(&output, &native_stat);
  double native_timeout = (double)timeout.secs + ((double)timeout.nanos / 1000000000.0);
  fuse_reply_attr(req, &native_stat, native_timeout);
}

static void bridge_readlink(fuse_req_t req, fuse_ino_t ino) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->readlink == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLBuf buf = {0};
  int err = ops->readlink(ops->userdata, req, ino, &buf);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  char *tmp = calloc(buf.len + 1, 1);
  if (tmp == NULL) {
    buf_release(&buf);
    fuse_reply_err(req, ENOMEM);
    return;
  }
  if (buf.len > 0 && buf.data != NULL) {
    memcpy(tmp, buf.data, buf.len);
  }
  fuse_reply_readlink(req, tmp);
  free(tmp);
  buf_release(&buf);
}

static void bridge_mknod(
    fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, dev_t rdev) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->mknod == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLEntry entry = {0};
  int err = ops->mknod(ops->userdata, req, parent, name, mode, rdev, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  fuse_reply_entry(req, &native_entry);
}

static void bridge_mkdir(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->mkdir == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLEntry entry = {0};
  int err = ops->mkdir(ops->userdata, req, parent, name, mode, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  fuse_reply_entry(req, &native_entry);
}

static void bridge_unlink(fuse_req_t req, fuse_ino_t parent, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->unlink != NULL ? ops->unlink(ops->userdata, req, parent, name) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_rmdir(fuse_req_t req, fuse_ino_t parent, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->rmdir != NULL ? ops->rmdir(ops->userdata, req, parent, name) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_symlink(fuse_req_t req, const char *link, fuse_ino_t parent, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->symlink == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLEntry entry = {0};
  int err = ops->symlink(ops->userdata, req, link, parent, name, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  fuse_reply_entry(req, &native_entry);
}

#if defined(__APPLE__)
static void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                          const char *newname) {
  uint32_t flags = 0;
#else
static void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                          const char *newname, unsigned int flags) {
#endif
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->rename != NULL
                ? ops->rename(ops->userdata, req, parent, name, newparent, newname, flags)
                : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_link(fuse_req_t req, fuse_ino_t ino, fuse_ino_t newparent, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->link == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLEntry entry = {0};
  int err = ops->link(ops->userdata, req, ino, newparent, name, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  fuse_reply_entry(req, &native_entry);
}

static void bridge_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->open == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->open(ops->userdata, req, ino, &info);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  file_info_to_native(&info, fi);
  if (fuse_reply_open(req, fi) == -ENOENT && ops->release != NULL) {
    FuseLLFileInfo reply_info;
    file_info_from_native(fi, &reply_info);
    ops->release(ops->userdata, req, ino, &reply_info);
  }
}

static void bridge_read(
    fuse_req_t req, fuse_ino_t ino, size_t size, off_t off, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->read == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  FuseLLBuf buf = {0};
  int err = ops->read(ops->userdata, req, ino, size, off, &info, &buf);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  fuse_reply_buf(req, (const char *)buf.data, buf.len);
  buf_release(&buf);
}

static void bridge_write(fuse_req_t req, fuse_ino_t ino, const char *buf, size_t size, off_t off,
                         struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->write == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  size_t written = 0;
  int err = ops->write(ops->userdata, req, ino, (const unsigned char *)buf, size, off, &info,
                       &written);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  fuse_reply_write(req, written);
}

static void bridge_flush(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->flush != NULL ? ops->flush(ops->userdata, req, ino, &info) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_release(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->release != NULL ? ops->release(ops->userdata, req, ino, &info) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_fsync(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->fsync != NULL ? ops->fsync(ops->userdata, req, ino, datasync, &info) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_opendir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->opendir != NULL ? ops->opendir(ops->userdata, req, ino, &info) : ENOSYS;
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  file_info_to_native(&info, fi);
  if (fuse_reply_open(req, fi) == -ENOENT && ops->releasedir != NULL) {
    FuseLLFileInfo reply_info;
    file_info_from_native(fi, &reply_info);
    ops->releasedir(ops->userdata, req, ino, &reply_info);
  }
}

static void bridge_readdir(
    fuse_req_t req, fuse_ino_t ino, size_t size, off_t off, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->readdir == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  FuseLLDirBuf db;
  db.req = req;
  db.size = size < 4096 ? 4096 : size;
  db.buf = malloc(db.size);
  db.offset = 0;
  if (db.buf == NULL) {
    fuse_reply_err(req, ENOMEM);
    return;
  }
  int err = ops->readdir(ops->userdata, req, ino, &info, off, size, &db);
  if (err != 0) {
    free(db.buf);
    fuse_reply_err(req, err);
    return;
  }
  fuse_reply_buf(req, db.buf, db.offset);
  free(db.buf);
}

static void bridge_releasedir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err = ops->releasedir != NULL ? ops->releasedir(ops->userdata, req, ino, &info) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_fsyncdir(
    fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  int err =
      ops->fsyncdir != NULL ? ops->fsyncdir(ops->userdata, req, ino, datasync, &info) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_statfs(fuse_req_t req, fuse_ino_t ino) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->statfs == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLStatVfs stat = {0};
  int err = ops->statfs(ops->userdata, req, ino, &stat);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  struct statvfs native_stat = EMPTY_STATVFS;
  statvfs_to_native(&stat, &native_stat);
  fuse_reply_statfs(req, &native_stat);
}

#if defined(__APPLE__)
static void bridge_setxattr(
    fuse_req_t req, fuse_ino_t ino, const char *name, const char *value, size_t size, int flags,
    uint32_t position) {
  if (position != 0) {
    fuse_reply_err(req, EPERM);
    return;
  }
#else
static void bridge_setxattr(
    fuse_req_t req, fuse_ino_t ino, const char *name, const char *value, size_t size, int flags) {
#endif
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->setxattr != NULL
                ? ops->setxattr(ops->userdata, req, ino, name, (const unsigned char *)value, size,
                                flags)
                : ENOSYS;
  fuse_reply_err(req, err);
}

#if defined(__APPLE__)
static void bridge_getxattr(
    fuse_req_t req, fuse_ino_t ino, const char *name, size_t size, uint32_t position) {
  if (position != 0) {
    fuse_reply_err(req, EPERM);
    return;
  }
#else
static void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size) {
#endif
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->getxattr == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLBuf buf = {0};
  int err = ops->getxattr(ops->userdata, req, ino, name, &buf);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  if (size == 0) {
    fuse_reply_xattr(req, buf.len);
    buf_release(&buf);
    return;
  }
  if (buf.len > size) {
    buf_release(&buf);
    fuse_reply_err(req, ERANGE);
    return;
  }
  fuse_reply_buf(req, (const char *)buf.data, buf.len);
  buf_release(&buf);
}

static void bridge_listxattr(fuse_req_t req, fuse_ino_t ino, size_t size) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->listxattr == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLBuf buf = {0};
  int err = ops->listxattr(ops->userdata, req, ino, &buf);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  if (size == 0) {
    fuse_reply_xattr(req, buf.len);
    buf_release(&buf);
    return;
  }
  if (buf.len > size) {
    buf_release(&buf);
    fuse_reply_err(req, ERANGE);
    return;
  }
  fuse_reply_buf(req, (const char *)buf.data, buf.len);
  buf_release(&buf);
}

static void bridge_removexattr(fuse_req_t req, fuse_ino_t ino, const char *name) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->removexattr != NULL ? ops->removexattr(ops->userdata, req, ino, name) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_access(fuse_req_t req, fuse_ino_t ino, int mask) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  int err = ops->access != NULL ? ops->access(ops->userdata, req, ino, mask) : ENOSYS;
  fuse_reply_err(req, err);
}

static void bridge_create(
    fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, struct fuse_file_info *fi) {
  FuseLLSession *session = (FuseLLSession *)fuse_req_userdata(req);
  FuseLLOps *ops = session_ops(session);
  if (ops->create == NULL) {
    fuse_reply_err(req, ENOSYS);
    return;
  }
  FuseLLFileInfo info;
  file_info_from_native(fi, &info);
  FuseLLEntry entry = {0};
  int err = ops->create(ops->userdata, req, parent, name, mode, &info, &entry);
  if (err != 0) {
    fuse_reply_err(req, err);
    return;
  }
  file_info_to_native(&info, fi);
  struct fuse_entry_param native_entry = EMPTY_ENTRY;
  entry_to_native(&entry, &native_entry);
  fuse_reply_create(req, &native_entry, fi);
}

static struct fuse_lowlevel_ops BRIDGE_OPS = {
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
};

FuseLLSession *fuse_ll_session_mount(int argc, const char *const *argv, const FuseLLOps *ops) {
  FuseLLSession *session = calloc(1, sizeof(FuseLLSession));
  if (session == NULL) {
    return NULL;
  }
  session->ops = *ops;
  session->args = (struct fuse_args)FUSE_ARGS_INIT(argc, (char **)argv);

#if defined(__APPLE__)
  if (fuse_parse_cmdline(&session->args, &session->mountpoint, NULL, NULL) == -1) {
    free(session);
    return NULL;
  }
  session->channel = fuse_mount(session->mountpoint, &session->args);
  if (session->channel == NULL) {
    fuse_opt_free_args(&session->args);
    free(session->mountpoint);
    free(session);
    return NULL;
  }
  session->session =
      fuse_lowlevel_new(&session->args, &BRIDGE_OPS, sizeof(BRIDGE_OPS), session);
  if (session->session == NULL) {
    fuse_unmount(session->mountpoint, session->channel);
    fuse_opt_free_args(&session->args);
    free(session->mountpoint);
    free(session);
    return NULL;
  }
  if (fuse_set_signal_handlers(session->session) != -1) {
    session->signal_handlers = 1;
  }
  fuse_session_add_chan(session->session, session->channel);
  session->mounted = 1;
#else
  if (fuse_parse_cmdline(&session->args, &session->cmdline) != 0) {
    free(session);
    return NULL;
  }
  if (session->cmdline.mountpoint == NULL) {
    free(session);
    return NULL;
  }
  session->mountpoint = session->cmdline.mountpoint;
  session->session = fuse_session_new(&session->args, &BRIDGE_OPS, sizeof(BRIDGE_OPS), session);
  if (session->session == NULL) {
    free(session->mountpoint);
    fuse_opt_free_args(&session->args);
    free(session);
    return NULL;
  }
  if (fuse_set_signal_handlers(session->session) == 0) {
    session->signal_handlers = 1;
  }
  if (fuse_session_mount(session->session, session->mountpoint) != 0) {
    if (session->signal_handlers) {
      fuse_remove_signal_handlers(session->session);
    }
    fuse_session_destroy(session->session);
    free(session->mountpoint);
    fuse_opt_free_args(&session->args);
    free(session);
    return NULL;
  }
  session->mounted = 1;
#endif

  return session;
}

int fuse_ll_session_run(FuseLLSession *session) {
  if (session == NULL || session->session == NULL) {
    return 1;
  }
#if defined(__APPLE__)
  return fuse_session_loop(session->session);
#else
  fuse_daemonize(session->cmdline.foreground);
  if (session->cmdline.singlethread) {
    return fuse_session_loop(session->session);
  }
  struct fuse_loop_config config = {
      .clone_fd = session->cmdline.clone_fd,
      .max_idle_threads = session->cmdline.max_idle_threads,
  };
  return fuse_session_loop_mt(session->session, &config);
#endif
}

void fuse_ll_session_unmount(FuseLLSession *session) {
  if (session == NULL || !session->mounted) {
    return;
  }
#if defined(__APPLE__)
  if (session->session != NULL && session->channel != NULL) {
    fuse_session_remove_chan(session->channel);
  }
  if (session->mountpoint != NULL && session->channel != NULL) {
    fuse_unmount(session->mountpoint, session->channel);
  }
#else
  if (session->session != NULL) {
    fuse_session_unmount(session->session);
  }
#endif
  session->mounted = 0;
}

void fuse_ll_session_destroy(FuseLLSession *session) {
  if (session == NULL) {
    return;
  }
  fuse_ll_session_unmount(session);
  if (session->signal_handlers && session->session != NULL) {
    fuse_remove_signal_handlers(session->session);
  }
  if (session->session != NULL) {
    fuse_session_destroy(session->session);
  }
  if (session->mountpoint != NULL) {
    free(session->mountpoint);
  }
  fuse_opt_free_args(&session->args);
  free(session);
}
