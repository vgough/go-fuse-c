#ifndef _WRAPPER_H_
#define _WRAPPER_H_

#define FUSE_USE_VERSION 29
#define _FILE_OFFSET_BITS 64

#include <fuse/fuse_lowlevel.h>

int MountAndRun(void *userdata, int argc, char *argv[]);

#endif  // _WRAPPER_H_
