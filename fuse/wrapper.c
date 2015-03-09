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

static const struct stat emptyStat;

int MountAndRun(int id, int argc, char *argv[], const struct fuse_lowlevel_ops *ops) {
  struct fuse_args args = FUSE_ARGS_INIT(argc, argv);
  struct fuse_chan *ch;
  char *mountpoint;
  int err = -1;

  if (fuse_parse_cmdline(&args, &mountpoint, NULL, NULL) != -1 &&
      (ch = fuse_mount(mountpoint, &args)) != NULL) {
    struct fuse_session *se;

    se = fuse_lowlevel_new(&args, ops, sizeof(struct fuse_lowlevel_ops), &id);
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

void fill_timespec(struct timespec *out, time_t sec, unsigned long nsec) {
  out->tv_sec = sec;
  out->tv_nsec = nsec;
}
