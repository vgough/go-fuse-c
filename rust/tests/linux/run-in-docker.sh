#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "this runner requires a Linux host with Docker and /dev/fuse" >&2
  exit 1
fi

if [[ ! -e /dev/fuse ]]; then
  echo "/dev/fuse is required on the host to pass through to Docker" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
image_name="${IMAGE_NAME:-go-fuse-c-rust-linux-test}"

docker build -t "${image_name}" -f "${repo_root}/rust/tests/linux/Dockerfile" "${repo_root}"
docker run --rm \
  --cap-add SYS_ADMIN \
  --device /dev/fuse \
  --security-opt apparmor=unconfined \
  --security-opt seccomp=unconfined \
  "${image_name}"
