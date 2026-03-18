FROM rust:1.94-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        clang \
        fuse3 \
        libfuse3-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY rust /workspace/rust

WORKDIR /workspace/rust

ENTRYPOINT ["bash", "/workspace/rust/tests/linux/smoke.sh"]
