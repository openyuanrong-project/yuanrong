go_packages = [
    "runtime_go",
]

def yr_go_test(name, sanitizer):
    asan_options=""
    cgo_ldflags=""
    coverage_suffix = "_coverage.out"
    test_result_out = "test_result.out"
    if sanitizer != "off":
        cgo_ldflags = "-fsanitize={}".format(sanitizer)
        coverage_suffix = "_coverage_{}.out".format(sanitizer)
        test_result_out = "test_result_{}.out".format(sanitizer)
    if sanitizer == "address":
        asan_options= "detect_odr_violation=0"

    native.genrule(
        name = name,
        srcs = [
            "//api/go/libruntime/cpplibruntime:libcpplibruntime.so",
            "@datasystem_sdk//:shared",
            "@metrics_sdk//:shared",
            ":go_sources",
        ],
        outs = [package + coverage_suffix for package in go_packages] + [test_result_out],
        cmd = r"""
            BASE_DIR="$$(pwd)" &&
            TEST_RESULT_OUT=$$BASE_DIR/$(location {test_result_out}) &&
            CGO_LINKDIR=$$(dirname $$BASE_DIR/$(location //api/go/libruntime/cpplibruntime:libcpplibruntime.so)) &&
            DATASYSTEM_DIR=$$(dirname $$BASE_DIR/$(locations @datasystem_sdk//:shared) | head -1) &&
            METRICS_DIR=$$(dirname $$BASE_DIR/$(locations @metrics_sdk//:shared) | head -1) &&
            export LD_LIBRARY_PATH=$$LD_LIBRARY_PATH:$$DATASYSTEM_DIR:$$METRICS_DIR:$$CGO_LINKDIR &&
            export CGO_LDFLAGS="{cgo_ldflags} -L$$CGO_LINKDIR -lcpplibruntime" &&
            export ASAN_OPTIONS={asan_options} &&
            OUT_DIR=$$(dirname $$TEST_RESULT_OUT) &&
            for package in {all_packages}; do
                cd $$(realpath $$BASE_DIR/api/go) &&
                go mod tidy &&
                go test ./libruntime/... ./faassdk/... ./posixsdk/... ./yr/... -covermode=set -coverprofile=$$OUT_DIR/"$$package"{suffix} -cover -gcflags=all=-l -vet=off || exit 2
            done &&
            echo "$$OUT_DIR" > $$TEST_RESULT_OUT
        """.format(
            cgo_ldflags = cgo_ldflags,
            asan_options = asan_options,
            test_result_out = test_result_out,
            all_packages = " ".join(go_packages),
            suffix = coverage_suffix,
        ),
        local = True,
        visibility = ["//visibility:public"],
    )