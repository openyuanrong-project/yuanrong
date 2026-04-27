load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def preload_opentelemetry():
    http_archive(
        name = "com_github_opentelemetry_proto",
        sha256 = "df491a05f3fcbf86cc5ba5c9de81f6a624d74d4773d7009d573e37d6e2b6af64",
        strip_prefix = "opentelemetry-proto-1.1.0",
        urls = ["https://github.com/open-telemetry/opentelemetry-proto/archive/refs/tags/v1.1.0.tar.gz"],
        build_file = "@opentelemetry_cpp//bazel:opentelemetry_proto.BUILD",
    )

    http_archive(
        name = "com_github_jupp0r_prometheus_cpp",
        sha256 = "8104d3b216aae60a1d0bca04adea4ba9ac1748eb1ed8646e123cf8e1591d99a3",
        strip_prefix = "prometheus-cpp-1.1.0",
        urls = ["https://github.com/jupp0r/prometheus-cpp/archive/refs/tags/v1.1.0.zip"],
    )

    http_archive(
        name = "github_nlohmann_json",
        build_file = "@opentelemetry_cpp//bazel:nlohmann_json.BUILD",
        sha256 = "0deac294b2c96c593d0b7c0fb2385a2f4594e8053a36c52b11445ef4b9defebb",
        strip_prefix = "nlohmann-json-v3.11.3",
        urls = ["https://gitee.com/mirrors/nlohmann-json/repository/archive/v3.11.3.zip"],
    )


