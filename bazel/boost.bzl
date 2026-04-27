cc_library(
    name = "boost",
    srcs = glob(["lib/*.a","lib/*.so*"]),
    hdrs = glob(["boost/**/*.hpp", "boost/**/*.h", "boost/**/*.ipp"]),
    includes = ["."],
    linkopts = ["-L$(GENDIR)/external/boost/lib"],
    visibility = ["//visibility:public"],
)