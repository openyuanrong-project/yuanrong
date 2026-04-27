load("@com_google_googleapis//:repository_rules.bzl", "switched_rules_by_language")
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
def grpc_extra_deps():
    protobuf_deps()
    switched_rules_by_language(
        name = "com_google_googleapis_imports",
        python = True,
        cc = True,
        grpc = True,
    )