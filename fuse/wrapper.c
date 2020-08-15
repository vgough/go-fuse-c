#include "bridge.h"

#if defined(__APPLE__)
#include <osxfuse/fuse/fuse_common.h>  // for fuse_mount, etc
#include <osxfuse/fuse/fuse_opt.h>     // for fuse_opt_free_args, etc
#else
#include <fuse/fuse_common.h>  // for fuse_mount, etc
#include <fuse/fuse_opt.h>     // for fuse_opt_free_args, etc
#endif

#include <stdio.h>      // for NULL
#include <sys/stat.h>   // for stat
#include <sys/types.h>  // for off_t
#include <unistd.h>     // for getgid, getuid

#include "_cgo_export.h"  // IWYU pragma: keep

static const struct stat emptyStat;

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

struct fuse_args *InitArgs(int argc, char *argv[]) {
  struct fuse_args *args = malloc(sizeof(struct fuse_args));

  args->argc = argc;
  args->argv = argv;
  args->allocated = 0;

  return args;
}

char *ParseMountpoint(struct fuse_args *args) {
  char *mountpoint;

  if (fuse_parse_cmdline(args, &mountpoint, NULL, NULL) == -1) {
    printf("unable to parse cmdline\n");
    return NULL;
  }

  if (mountpoint == NULL) {
    fprintf(stderr, "no mount point specified\n");
    return NULL;
  }

  return mountpoint;
}

struct fuse_chan *Mount(char *mountpoint, struct fuse_args *args) {
  struct fuse_chan *ch;

  ch = fuse_mount(mountpoint, args);
  if (ch == NULL) {
    fuse_opt_free_args(args);
    return NULL;
  }

  return ch;
}

struct fuse_session *NewSession(char *mountpoint, struct fuse_args *args, struct fuse_chan *ch) {
  struct fuse_session *se;

  se = fuse_lowlevel_new(args, &bridge_ll_ops, sizeof(struct fuse_lowlevel_ops), mountpoint);
  if (se == NULL) {
    return NULL;
  }

  if (fuse_set_signal_handlers(se) == -1) {
    fuse_session_destroy(se);
    return NULL;
  }

  fuse_session_add_chan(se, ch);

  return se;
}

int Run(char *mountpoint, struct fuse_session *se, struct fuse_chan *ch) {
  int err = -1;

  err = fuse_session_loop(se);

  fuse_remove_signal_handlers(se);
  fuse_session_remove_chan(ch);
  
  fuse_session_destroy(se);

  return err ? 1 : 0;
}

void Exit(char *mountpoint, struct fuse_session *se, struct fuse_chan *ch) {
  fuse_session_exit(se);

  fuse_unmount(mountpoint, ch);
}

// Returns 0 on success.
int DirBufAdd(struct DirBuf *db, const char *name, fuse_ino_t ino, int mode, off_t next) {
  struct stat stbuf = emptyStat;
  stbuf.st_ino = ino;
  stbuf.st_mode = mode;
  stbuf.st_uid = getuid();
  stbuf.st_gid = getgid();

  char *buf = db->buf + db->offset;
  size_t left = db->size - db->offset;
  size_t size = fuse_add_direntry(db->req, buf, left, name, &stbuf, next);
  if (size < left) {
    db->offset += size;
    return 0;
  }

  return 1;
}

void FillTimespec(struct timespec *out, time_t sec, unsigned long nsec) {
  out->tv_sec = sec;
  out->tv_nsec = nsec;
}
