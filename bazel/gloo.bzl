load("@yuanrong_multi_language_runtime//bazel:yr.bzl", "filter_files_with_suffix")

cc_import(
    name = "gloo",
    hdrs = glob(["include/gloo/**/*.h"]),
    shared_library = "lib/libgloo.so",
    visibility = ["//visibility:public"],
)

filter_files_with_suffix(
    name = "shared",
    srcs = glob(["lib/lib*.so*"]),
    suffix = ".so",
    visibility = ["//visibility:public"],
)