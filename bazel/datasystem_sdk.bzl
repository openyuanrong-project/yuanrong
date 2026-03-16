load("@yuanrong_multi_language_runtime//bazel:yr.bzl", "filter_files_with_suffix")

cc_library(
    name = "lib_datasystem_sdk",
    srcs = [
        "lib/libdatasystem.so",
        "lib/libsecurec.so",
        "lib/libzmq.so.5",
        "lib/libprotobuf.so.25.5.0",
        "lib/librpc_option_protos.so",
        "lib/libabseil_dll.so.2407.0.0",
        "lib/libcommon_flags.so",
        "lib/libetcdapi_proto.so",
        "lib/libgrpc++.so.1.65",
        "lib/libgrpc.so.42",
        "lib/libupb_json_lib.so.42",
        "lib/libupb_textformat_lib.so.42",
        "lib/libutf8_range_lib.so.42",
        "lib/libupb_message_lib.so.42",
        "lib/libupb_base_lib.so.42",
        "lib/libupb_mem_lib.so.42",
        "lib/libgpr.so.42",
    ] + glob([
        "lib/libds-spdlog.so.*",
        "lib/libtbb.so*",
    ]),
    hdrs = glob(["include/**/*.h"]),
    strip_include_prefix = "include",
    visibility = ["//visibility:public"],
    alwayslink = True,
)

filter_files_with_suffix(
    name = "shared",
    srcs = glob(["lib/lib*.so*"]),
    suffix = ".so",
    visibility = ["//visibility:public"],
)
