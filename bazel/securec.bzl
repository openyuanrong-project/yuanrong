cc_library(
    name = "securec",
    srcs = glob([
        "src/*.c",
        "src/*.h", 
        "src/*.inl",
    ]),
    hdrs = glob(["include/*.h"]),
    strip_include_prefix = "include",
    copts = [],
    visibility = ["//visibility:public"],
)