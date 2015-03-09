
#include "wrapper.h"

// Methods to create and work with a fuse_req test instance.
// This provides a fuse_req_t during unit tests.
// Note that fuse_req is not a public type, so the contents are opaque.
fuse_req_t new_fuse_test_req(int id, int userdata);
void free_fuse_test_req(fuse_req_t req);
int fuse_test_req_id(fuse_req_t req);

// Bridge functions.  During production use, these are called by the FUSE lowlevel library.
// The signatures are provided so that they can be called by unit tests, bypassing FUSE.
void bridge_init(void *userdata, struct fuse_conn_info *conn);
void bridge_destroy(void *userdata);
void bridge_lookup(fuse_req_t req, fuse_ino_t parent, const char *name);
void bridge_forget(fuse_req_t req, fuse_ino_t ino, unsigned long nlookup);
void bridge_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_setattr(fuse_req_t req, fuse_ino_t ino, struct stat *attr, int to_set,
                    struct fuse_file_info *fi);
void bridge_readlink(fuse_req_t req, fuse_ino_t ino);
void bridge_mknod(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode, dev_t rdev);
void bridge_mkdir(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode);
void bridge_unlink(fuse_req_t req, fuse_ino_t parent, const char *name);
void bridge_rmdir(fuse_req_t req, fuse_ino_t parent, const char *name);
void bridge_symlink(fuse_req_t req, const char *link, fuse_ino_t parent, const char *name);
void bridge_rename(fuse_req_t req, fuse_ino_t parent, const char *name, fuse_ino_t newparent,
                   const char *newname);
void bridge_link(fuse_req_t req, fuse_ino_t ino, fuse_ino_t newparent, const char *newname);
void bridge_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_read(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off, struct fuse_file_info *fi);
void bridge_write(fuse_req_t req, fuse_ino_t ino, const char *buf, size_t size, off_t off,
                  struct fuse_file_info *fi);
void bridge_flush(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_release(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_fsync(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi);
void bridge_opendir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_readdir(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off,
                    struct fuse_file_info *fi);
void bridge_releasedir(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void bridge_fsyncdir(fuse_req_t req, fuse_ino_t ino, int datasync, struct fuse_file_info *fi);
void bridge_statfs(fuse_req_t req, fuse_ino_t ino);

#ifdef __APPLE__
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags, uint32_t position);
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size,
                     uint32_t position);
#else
void bridge_setxattr(fuse_req_t req, fuse_ino_t ino, const char *name, const char *value,
                     size_t size, int flags);
void bridge_getxattr(fuse_req_t req, fuse_ino_t ino, const char *name, size_t size);
#endif

void bridge_listxattr(fuse_req_t req, fuse_ino_t ino, size_t size);
void bridge_removexattr(fuse_req_t req, fuse_ino_t ino, const char *name);
void bridge_access(fuse_req_t req, fuse_ino_t ino, int mask);
void bridge_create(fuse_req_t req, fuse_ino_t parent, const char *name, mode_t mode,
                   struct fuse_file_info *fi);
void bridge_getlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock);
void bridge_setlk(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, struct flock *lock,
                  int sleep);
void bridge_bmap(fuse_req_t req, fuse_ino_t ino, size_t blocksize, uint64_t idx);
void bridge_ioctl(fuse_req_t req, fuse_ino_t ino, int cmd, void *arg, struct fuse_file_info *fi,
                  unsigned flags, const void *in_buf, size_t in_bufsz, size_t out_bufsz);

void bridge_flock(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi, int op);
void bridge_fallocate(fuse_req_t req, fuse_ino_t ino, int mode, off_t offset, off_t length,
                      struct fuse_file_info *fi);
