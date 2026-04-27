cc_library(
    name = "nlohmann_json",
    srcs = [],
    hdrs = ["single_include/nlohmann/json.hpp"],
    strip_include_prefix = "single_include/nlohmann/",
    copts = [],
    visibility = ["//visibility:public"],
)