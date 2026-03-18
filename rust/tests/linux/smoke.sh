#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "linux smoke test must run on Linux" >&2
  exit 1
fi

if [[ ! -e /dev/fuse ]]; then
  echo "/dev/fuse is required inside the container" >&2
  exit 1
fi

mount_dir="/tmp/fuse-lowlevel-hello-mount"
log_file="/tmp/fuse-lowlevel-hello.log"
hello_pid=""

cleanup() {
  if mountpoint -q "${mount_dir}" 2>/dev/null; then
    fusermount3 -u "${mount_dir}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${hello_pid}" ]] && kill -0 "${hello_pid}" 2>/dev/null; then
    kill "${hello_pid}" >/dev/null 2>&1 || true
    wait "${hello_pid}" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT

rm -rf "${mount_dir}"
mkdir -p "${mount_dir}"

cargo build -p fuse-lowlevel --example hello

./target/debug/examples/hello "${mount_dir}" -f >"${log_file}" 2>&1 &
hello_pid="$!"

for _ in $(seq 1 100); do
  if mountpoint -q "${mount_dir}"; then
    break
  fi
  if ! kill -0 "${hello_pid}" 2>/dev/null; then
    echo "hello example exited before the mount became active" >&2
    cat "${log_file}" >&2 || true
    exit 1
  fi
  sleep 0.1
done

if ! mountpoint -q "${mount_dir}"; then
  echo "mount never became active" >&2
  cat "${log_file}" >&2 || true
  exit 1
fi

if [[ ! -f "${mount_dir}/hello" ]]; then
  echo "mounted filesystem does not expose /hello" >&2
  ls -la "${mount_dir}" >&2 || true
  exit 1
fi

actual="$(cat "${mount_dir}/hello")"
if [[ "${actual}" != "Hello World!" ]]; then
  echo "unexpected file contents from mounted hello example" >&2
  printf 'expected: %q\n' "Hello World!" >&2
  printf 'actual: %q\n' "${actual}" >&2
  exit 1
fi

fusermount3 -u "${mount_dir}"
wait "${hello_pid}"
