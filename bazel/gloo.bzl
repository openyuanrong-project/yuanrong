load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_library")
load("@yuanrong_multi_language_runtime//bazel:yr.bzl", "filter_files_with_suffix")

# Files to exclude from the build
GLOO_EXCLUDES = [
    "gloo/benchmark/**",
    "gloo/examples/**",
    "gloo/test/**",
    "gloo/cuda/**",
    "gloo/cuda*",
    "gloo/hip/**",
    "gloo/nccl/**",
    "gloo/mpi/**",
    "gloo/rendezvous/redis_store*",
    "gloo/transport/uv/**",
    "gloo/transport/tcp/tls/**",
    "gloo/common/win.*",
]

# All source files - use glob to include all .cc files except excluded ones
GLOO_ALL_SRCS = glob(
    ["gloo/**/*.cc"],
    exclude = GLOO_EXCLUDES,
)

# All header files - use glob to include all .h files except excluded ones
GLOO_ALL_HDRS = glob(
    ["gloo/**/*.h"],
    exclude = GLOO_EXCLUDES,
)

genrule(
    name = "config_header",
    outs = ["gloo/config.h"],
    cmd = """
        cat <<'EOF' > $@
#pragma once

#define GLOO_VERSION_MAJOR 0
#define GLOO_VERSION_MINOR 5
#define GLOO_VERSION_PATCH 0

static_assert(GLOO_VERSION_MINOR < 100, "Invalid GLOO minor version");
static_assert(GLOO_VERSION_PATCH < 100, "Invalid GLOO patch version");

#define GLOO_VERSION (GLOO_VERSION_MAJOR * 10000 + GLOO_VERSION_MINOR * 100 + GLOO_VERSION_PATCH)

#define GLOO_USE_CUDA 0
#define GLOO_USE_NCCL 0
#define GLOO_USE_ROCM 0
#define GLOO_USE_RCCL 0
#define GLOO_USE_REDIS 0
#define GLOO_USE_IBVERBS 1
#define GLOO_USE_MPI 0
#define GLOO_USE_AVX 0
#define GLOO_USE_LIBUV 0

#define GLOO_HAVE_TRANSPORT_TCP 1
#define GLOO_HAVE_TRANSPORT_TCP_TLS 0
#define GLOO_HAVE_TRANSPORT_IBVERBS 1
#define GLOO_HAVE_TRANSPORT_UV 0

#define GLOO_USE_TORCH_DTYPES 0
EOF
    """,
)

cc_library(
    name = "gloo",
    srcs = GLOO_ALL_SRCS,
    hdrs = GLOO_ALL_HDRS + [
        ":config_header",
    ],
    includes = ["."],
    copts = [
        "-std=c++17",
        "-fPIC",
    ],
    linkopts = [
        "-lpthread",
        "-libverbs",
    ],
    linkstatic = True,
    visibility = ["//visibility:public"],
)