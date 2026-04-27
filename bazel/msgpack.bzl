cc_library(
    name = "msgpack",
    srcs = [],
    hdrs = glob(["include/**/*.h", "include/**/*.hpp"]),
    strip_include_prefix = "include",
    copts = [],
    deps = [
        "@boost//:boost",
    ],
    visibility = ["//visibility:public"],
)