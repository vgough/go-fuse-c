#include "bridge.h"

#if defined(__APPLE__)
#include <osxfuse/fuse/fuse_common.h>  // for fuse_mount, etc
#include <osxfuse/fuse/fuse_opt.h>     // for fuse_opt_free_args, etc
#else
#include <fuse_common.h>
#include <fuse_lowlevel.h>
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

int MountAndRun(int id, int argc, char *argv[]) {
  struct fuse_args args = FUSE_ARGS_INIT(argc, argv);
  struct fuse_session *se;
  struct fuse_cmdline_opts opts;
  struct fuse_loop_config config;
  int ret = -1;

  if (fuse_parse_cmdline(&args, &opts) != 0) {
    return 1;
  }
  if (opts.show_help) {
    printf("usage: %s [options] <mountpoint>\n\n", argv[0]);
    fuse_cmdline_help();
    fuse_lowlevel_help();
    ret = 0;
    goto err_out1;
  } else if (opts.show_version) {
    printf("FUSE library version %s\n", fuse_pkgversion());
    fuse_lowlevel_version();
    ret = 0;
    goto err_out1;
  }

  if (opts.mountpoint == NULL) {
    printf("usage: %s [options] <mountpoint>\n", argv[0]);
    printf("       %s --help\n", argv[0]);
    ret = 1;
    goto err_out1;
  }

  se = fuse_session_new(&args, &bridge_ll_ops, sizeof(bridge_ll_ops), &id);
  if (se == NULL) goto err_out1;

  if (fuse_set_signal_handlers(se) != 0) goto err_out2;

  if (fuse_session_mount(se, opts.mountpoint) != 0) goto err_out3;

  fuse_daemonize(opts.foreground);

  /* Block until ctrl+c or fusermount -u */
  if (opts.singlethread)
    ret = fuse_session_loop(se);
  else {
    config.clone_fd = opts.clone_fd;
    config.max_idle_threads = opts.max_idle_threads;
    ret = fuse_session_loop_mt(se, &config);
  }

  fuse_session_unmount(se);
err_out3:
  fuse_remove_signal_handlers(se);
err_out2:
  fuse_session_destroy(se);
err_out1:
  free(opts.mountpoint);
  fuse_opt_free_args(&args);

  return ret ? 1 : 0;
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
