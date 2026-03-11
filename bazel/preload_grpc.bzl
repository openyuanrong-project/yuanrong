load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//bazel:grpc_upb_repository.bzl", "grpc_upb_repository")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("//bazel:local_patched_repository.bzl", "local_patched_repository")

def preload_grpc():
    http_archive(
        name = "com_google_absl",
        sha256 = "104dead3edd7b67ddeb70c37578245130d6118efad5dad4b618d7e26a5331f55",
        strip_prefix = "abseil-cpp-20240722.0",
        urls = [
            "https://gitee.com/mirrors/abseil-cpp/repository/archive/20240722.0.zip",
        ],
        patches = ["@yuanrong_multi_language_runtime//patch:absl_failure_signal_handler.patch"],
    )

    http_archive(
        name = "com_google_protobuf",
        strip_prefix = "protobuf_source-v3.25.5",
        sha256 = "4640cb69abb679e2a4b061dfeb7debb3170b592e4ac6e3f16dbaaa4aac0710bd",
        urls = ["https://gitee.com/mirrors/protobuf_source/repository/archive/v3.25.5.zip"],
        patches = ["@yuanrong_multi_language_runtime//patch:protobuf_gcc_7_3.patch"],
    )

    http_archive(
        name = "utf8_range",
        strip_prefix = "utf8_range-d863bc33e15cba6d873c878dcca9e6fe52b2f8cb",
        sha256 = "568988b5f7261ca181468dba38849fabf59dd9200fb2ed4b2823da187ef84d8c",
        urls = ["https://github.com/protocolbuffers/utf8_range/archive/d863bc33e15cba6d873c878dcca9e6fe52b2f8cb.zip"],
    )

    http_archive(
        name = "cython",
        url = "https://gitee.com/mirrors/cython/repository/archive/3.0.10.zip",
        sha256 = "e7fd54afdfef123be52a63e17e46eec2942f6a8012c97030dc68e6e10ed16f13",
        strip_prefix = "cython-3.0.10",
        build_file = "@com_github_grpc_grpc//third_party:cython.BUILD",
    )

    http_archive(
        name = "zlib",
        strip_prefix = "zlib-v1.3.1",
        urls = ["https://gitee.com/mirrors/zlib/repository/archive/v1.3.1.zip"],
        sha256 = "7c31009abc4e76ddc32e1448b6051bafe5f606aac158bb36166100a21ec170c6",
        build_file = "@com_google_protobuf//:third_party/zlib.BUILD",
    )
    
    local_patched_repository(
        name = "com_github_grpc_grpc",
        path = "./thirdparty/grpc",
        patch_files = [
         "@yuanrong_multi_language_runtime//patch:grpc_1.65.patch",
         "@yuanrong_multi_language_runtime//patch:grpc_1_65_4_gcc_7_3.patch"
        ]
     )

    grpc_upb_repository(
        name = "upb",
        path = Label("@com_github_grpc_grpc//:WORKSPACE")
    )

    native.new_local_repository(
        name = "boringssl",
        build_file = "//bazel:openssl.bazel",
        path = "./thirdparty/openssl/"
    )

    http_archive(
        name = "com_googlesource_code_re2",
        url = "https://gitee.com/mirrors/re2/repository/archive/2024-02-01.zip",
        sha256 = "54bff0e995b101e1865dcea5d052ec10b3aadb6f8c57b5c03c9eeccddb00a08a",
        strip_prefix = "re2-2024-02-01",
    )

    http_archive(
        name = "com_google_googleapis",
        url = "https://github.com/googleapis/googleapis/archive/541b1ded4abadcc38e8178680b0677f65594ea6f.zip",
        sha256 = "7ebab01b06c555f4b6514453dc3e1667f810ef91d1d4d2d3aa29bb9fcb40a900",
        strip_prefix = "googleapis-541b1ded4abadcc38e8178680b0677f65594ea6f",
    )

    http_archive(
        name = "com_github_cares_cares",
        url = "https://gitee.com/mirrors/c-ares/repository/archive/cares-1_19_1.zip",
        sha256 = "edcaac184aff0e6b6eb7b9ede7a55f36c7fc04085d67fecff2434779155dd8ce",
        strip_prefix = "c-ares-cares-1_19_1",
    )
