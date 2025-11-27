workspace(name = "yuanrong_multi_language_runtime")

load("//bazel:hazel_workspace.bzl", "hw_rules")

hw_rules()

load("@rules_foreign_cc//foreign_cc:repositories.bzl", "rules_foreign_cc_dependencies")

rules_foreign_cc_dependencies()

load("@rules_python//python:repositories.bzl", "python_register_toolchains")
python_register_toolchains(
    name = "python3_9",
    ignore_root_user_error = True,
    python_version = "3.9.15",
    register_toolchains = False,
    register_coverage_tool = True,
)

load("@python3_9//:defs.bzl", python39 = "interpreter")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "new_git_repository")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//bazel:local_patched_repository.bzl", "local_patched_repository")

http_archive(
    name = "rules_jvm_external",
    sha256 = "b17d7388feb9bfa7f2fa09031b32707df529f26c91ab9e5d909eb1676badd9a6",
    strip_prefix = "rules_jvm_external-4.5",
    url = "https://github.com/bazel-contrib/rules_jvm_external/archive/refs/tags/4.5.zip",
)

load("@rules_jvm_external//:defs.bzl", "maven_install")

maven_install(
    artifacts = [
        "com.google.code.gson:gson:2.11.0",
        "org.apache.commons:commons-lang3:3.18.0",
        "org.apache.maven.plugins:maven-assembly-plugin:3.4.2",
        "org.apache.maven.plugins:maven-compiler-plugin:3.10.1",
        "commons-io:commons-io:2.16.1",
        "org.json:json:20230227",
        "org.msgpack:jackson-dataformat-msgpack:0.9.3",
        "org.msgpack:msgpack-core:0.9.3",
        "com.fasterxml.jackson.core:jackson-core:2.18.2",
        "com.fasterxml.jackson.core:jackson-databind:2.18.2",
        "org.apache.logging.log4j:log4j-slf4j-impl:2.23.1",
        "org.apache.logging.log4j:log4j-api:2.23.1",
        "org.apache.logging.log4j:log4j-core:2.23.1",
        "org.slf4j:slf4j-api:1.7.36",
        "org.powermock:powermock-module-junit4:2.0.4",
        "org.powermock:powermock-api-mockito2:2.0.4",
        "junit:junit:4.11",
        "org.jacoco:org.jacoco.agent:0.8.8",
        "org.projectlombok:lombok:1.18.36",
        "org.ow2.asm:asm:9.7",
    ],
    repositories = [
        "https://mirrors.huaweicloud.com/repository/maven/",
    ],
)

local_patched_repository(
    name = "spdlog",
    path = "../thirdparty/spdlog/",
    build_file = "@//bazel:spdlog.bzl",
    patch_files = [
         "@yuanrong_multi_language_runtime//patch:spdlog-change-namespace-and-library-name-with-yr.patch",
    ]
)

http_archive(
    name = "nlohmann_json",
    build_file = "@//bazel:nlohmann_json.bzl",
    sha256 = "0deac294b2c96c593d0b7c0fb2385a2f4594e8053a36c52b11445ef4b9defebb",
    strip_prefix = "nlohmann-json-v3.11.3",
    urls = ["https://gitee.com/mirrors/nlohmann-json/repository/archive/v3.11.3.zip"],
)

http_archive(
    name = "gtest",
    sha256 = "647924848ca7cb91ba5e34260132902886e1bd140428bd3bd7b4e8fa6c6c8904",
    strip_prefix = "googletest-v1.13.0",
    urls = ["https://gitee.com/mirrors/googletest/repository/archive/v1.13.0.zip"],
)

http_archive(
    name = "remote_coverage_tools",
    sha256 = "7006375f6756819b7013ca875eab70a541cf7d89142d9c511ed78ea4fefa38af",
    urls = [
        "https://mirror.bazel.build/bazel_coverage_output_generator/releases/coverage_output_generator-v2.6.zip",
    ],
)

load("//bazel:preload_grpc.bzl", "preload_grpc")

preload_grpc()

load("//bazel:preload_opentelemetry.bzl", "preload_opentelemetry")

preload_opentelemetry()

http_archive(
    name = "opentelemetry_cpp",
    sha256 = "7735cc56507149686e6019e06f588317099d4522480be5f38a2a09ec69af1706",
    strip_prefix = "opentelemetry-cpp-1.13.0",
    urls = ["https://github.com/open-telemetry/opentelemetry-cpp/archive/refs/tags/v1.13.0.tar.gz"],
)

load("@opentelemetry_cpp//bazel:repository.bzl", "opentelemetry_cpp_deps")
opentelemetry_cpp_deps()

load("@opentelemetry_cpp//bazel:extra_deps.bzl", "opentelemetry_extra_deps")
opentelemetry_extra_deps()

load("@com_github_grpc_grpc//bazel:grpc_deps.bzl", "grpc_deps")

grpc_deps()
load("//bazel:grpc_extra_deps.bzl", "grpc_extra_deps")
grpc_extra_deps()
new_local_repository(
    name = "boost",
    build_file = "@//bazel:boost.bzl",
    path = "../thirdparty/boost/",
)

http_archive(
    name = "msgpack",
    build_file = "@//bazel:msgpack.bzl",
    sha256 = "c51bcee3814b20d38a2e5bdf315e424344adb5330a64f4669cbbcd3cb6c89e27",
    strip_prefix = "msgpack-c-cpp-5.0.0",
    url = "https://gitee.com/mirrors/msgpack-c/repository/archive/cpp-5.0.0.zip",
)

http_archive(
    name = "yaml-cpp",
    sha256 = "6a05c681872d9465b8e2040b5211b1aa5cf30151dc4f3d7ed23ac75ce0fd9944",
    strip_prefix = "yaml-cpp-0.8.0",
    url = "https://gitee.com/mirrors/yaml-cpp/repository/archive/0.8.0.zip",
)

new_local_repository(
    name = "datasystem_sdk",
    build_file = "@//bazel:datasystem_sdk.bzl",
    path = "./datasystem/output/sdk/",
)

new_local_repository(
    name = "metrics_sdk",
    build_file = "@//bazel:metrics_sdk.bzl",
    path = "./metrics/",
)

load("@bazel_tools//tools/jdk:remote_java_repository.bzl", "remote_java_repository")

http_archive(
    name = "remote_java_tools",
    sha256 = "5cd59ea6bf938a1efc1e11ea562d37b39c82f76781211b7cd941a2346ea8484d",
    url = "https://mirror.bazel.build/bazel_java_tools/releases/java/v11.9/java_tools-v11.9.zip",
    patches = ["@yuanrong_multi_language_runtime//patch:remote_java_tools.patch"],
)

http_archive(
    name = "remote_java_tools_linux",
    sha256 = "512582cac5b7ea7974a77b0da4581b21f546c9478f206eedf54687eeac035989",
    urls = [
        "https://mirror.bazel.build/bazel_java_tools/releases/java/v11.9/java_tools_linux-v11.9.zip",
    ],
)

http_archive(
    name = "jacoco",
    sha256 = "6859d4deecc9fdd44f742bb8ff8e4ca71afca442cc8ce67aeb668dda951e8498",
    urls = ["https://github.com/jacoco/jacoco/releases/download/v0.8.8/jacoco-0.8.8.zip"],
    build_file = "@//bazel:jacoco.bzl"
)

local_patched_repository(
    name = "gloo",
    path = "../thirdparty/gloo",
    build_file = "@//bazel:gloo.bzl",
    patch_files = [
        "@//patch:gloo-fix-sign-compare.patch",
    ],
)

maybe(
    new_local_repository,
    name = "securec",
    build_file = "@//bazel:securec.bzl",
    path = "../thirdparty/libboundscheck",
)
