cc_library(
    name = "spdlog",
    srcs = [],
    hdrs = glob(["include/**/*.h", "include/**/*.hpp"]),
    strip_include_prefix = "include",
    copts = [],
    visibility = ["//visibility:public"],
)