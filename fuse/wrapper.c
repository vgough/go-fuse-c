#include "wrapper.h"

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

// bridge ops defined in bridge.c
extern struct fuse_lowlevel_ops bridge_ll_ops;
static const struct stat emptyStat;

int MountAndRun(int id, int argc, char *argv[]) {
  struct fuse_args args = FUSE_ARGS_INIT(argc, argv);
  struct fuse_chan *ch;
  char *mountpoint;
  int err = -1;

  if (fuse_parse_cmdline(&args, &mountpoint, NULL, NULL) != -1 &&
      (ch = fuse_mount(mountpoint, &args)) != NULL) {
    struct fuse_session *se;

    se = fuse_lowlevel_new(&args, &bridge_ll_ops, sizeof(bridge_ll_ops), &id);
    if (se != NULL) {
      if (fuse_set_signal_handlers(se) != -1) {
        fuse_session_add_chan(se, ch);
        err = fuse_session_loop(se);
        fuse_remove_signal_handlers(se);
        fuse_session_remove_chan(ch);
      }
      fuse_session_destroy(se);
    }
    fuse_unmount(mountpoint, ch);
  }
  fuse_opt_free_args(&args);

  return err ? 1 : 0;
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
