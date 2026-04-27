load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def hw_rules():

    # introduce the dependent packages.
    http_archive(
        name = "rules_java",
        sha256 = "469b7f3b580b4fcf8112f4d6d0d5a4ce8e1ad5e21fee67d8e8335d5f8b3debab",
        urls = [
            "https://github.com/bazelbuild/rules_java/releases/download/6.0.0/rules_java-6.0.0.tar.gz"
        ]
    )

    http_archive(
        name = "rules_ruby",
        sha256 = "347927fd8de6132099fcdc58e8f7eab7bde4eb2fd424546b9cd4f1c6f8f8bad8",
        strip_prefix = "rules_ruby-b7f3e9756f3c45527be27bc38840d5a1ba690436",
        urls = [
            "https://github.com/protocolbuffers/rules_ruby/archive/b7f3e9756f3c45527be27bc38840d5a1ba690436.zip"
        ]
    )

    http_archive(
        name = "bazel_skylib",
        sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
        urls = ["https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        ],
    )

    http_archive(
        name = "rules_cc",
        sha256 = "2037875b9a4456dce4a79d112a8ae885bbc4aad968e6587dca6e64f3a0900cdf",
        strip_prefix = "rules_cc-0.0.9",
        urls = [
            "https://github.com/bazelbuild/rules_cc/releases/download/0.0.9/rules_cc-0.0.9.tar.gz"
        ]
    )

    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "6dc2da7ab4cf5d7bfc7c949776b1b7c733f05e56edc4bcd9022bb249d2e2a996",
        urls = ["https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.39.1/rules_go-v0.39.1.zip",
        ],
    )

    http_archive(
        name = "rules_pkg",
        sha256 = "038f1caa773a7e35b3663865ffb003169c6a71dc995e39bf4815792f385d837d",
        urls = ["https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.4.0/rules_pkg-0.4.0.tar.gz",
        ],
    )

    http_archive(
        name = "rules_python",
        url = "https://github.com/bazelbuild/rules_python/archive/refs/tags/0.19.0.tar.gz",
        sha256 = "ffc7b877c95413c82bfd5482c017edcf759a6250d8b24e82f41f3c8b8d9e287e",
        strip_prefix = "rules_python-0.19.0",
    )

    http_archive(
        name = "build_bazel_rules_apple",
        url = "https://github.com/bazelbuild/rules_apple/archive/refs/tags/0.31.3.tar.gz",
        sha256 = "d6735ed25754dbcb4fce38e6d72c55b55f6afa91408e0b72f1357640b88bb49c",
        strip_prefix = "rules_apple-0.31.3",
    )

    http_archive(
        name = "build_bazel_rules_swift",
        url = "https://github.com/bazelbuild/rules_swift/archive/refs/tags/0.21.0.tar.gz",
        sha256 = "802c094df1642909833b59a9507ed5f118209cf96d13306219461827a00992da",
        strip_prefix = "rules_swift-0.21.0",
    )

    http_archive(
        name = "build_bazel_apple_support",
        url = "https://github.com/bazelbuild/apple_support/archive/refs/tags/0.10.0.tar.gz",
        sha256 = "c02a8c902f405e5ea12b815f426fbe429bc39a2628b290e50703d956d40f5542",
        strip_prefix = "apple_support-0.10.0",
    )

    http_archive(
        name = "rules_foreign_cc",
        url = "https://github.com/bazel-contrib/rules_foreign_cc/archive/refs/tags/0.9.0.tar.gz",
        sha256 = "2a4d07cd64b0719b39a7c12218a3e507672b82a97b98c6a89d38565894cf7c51",
        strip_prefix = "rules_foreign_cc-0.9.0",
    )