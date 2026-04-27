# BUILD file for etcd source

genrule(
    name = "bin",
    srcs = glob(["**/*"]),
    outs = ["etcd", "etcdctl", "etcdutl"],
    cmd = """
        set -e
        ETCD_DIR="$$(pwd)/external/etcd_source"
        OUTPUT_DIR="$$(pwd)/$(@D)"
        export GOTOOLCHAIN=local
        export GOPATH="$$(pwd)/go_cache"
        export GOMODCACHE="$$(pwd)/go_cache/mod"
        export GOCACHE="$$(pwd)/go_cache/cache"
        export GOFLAGS="-buildvcs=false"
        export HOME="$$(pwd)"
        mkdir -p "$$GOMODCACHE" "$$GOCACHE"
        cd "$$ETCD_DIR"
        bash build.sh
        cp -f "$$ETCD_DIR/bin/etcd" "$$OUTPUT_DIR/"
        cp -f "$$ETCD_DIR/bin/etcdctl" "$$OUTPUT_DIR/"
        cp -f "$$ETCD_DIR/bin/etcdutl" "$$OUTPUT_DIR/"
    """,
    visibility = ["//visibility:public"],
)
