load("@rules_cc//cc:find_cc_toolchain.bzl", "find_cc_toolchain")
COPTS = ["-DVERSION=1", "-Wno-stringop-overflow", "-Werror", "-fstack-protector-strong", "-Wno-deprecated-declarations", "-fPIC"]
COPTS_WITH_GLOO = COPTS + ["-DENABLE_GLOO"]
LOPTS = ["-DVERSION=1"]

def copy_file(name, srcs, dstdir = "", pre_cmd = "echo"):
    if dstdir.startswith("/"):
        fail("Subdirectory must be a relative path: " + dstdir)
    src_locations = " ".join(["$(locations {})".format(src) for src in srcs])
    native.genrule(
        name = name,
        srcs = srcs,
        outs = [name + ".out"],
        cmd = r"""
            mkdir -p -- {dstdir}
            for f in {locations}; do
                {pre_cmd} "$$f"
                rm -f -- {dstdir}$${{f##*/}}
                cp -af -- "$$f" {dstdir}
            done
            date > $@
        """.format(
            locations = src_locations,
            dstdir = "." + ("/" + dstdir).rstrip("/") + "/",
            pre_cmd = pre_cmd,
        ),
        local = 1,
        tags = ["no-cache"],
    )

def _filter_files_with_suffix_impl(ctx):
    suffix = ctx.attr.suffix
    filtered_files = [f for f in ctx.files.srcs if suffix in f.basename]
    return [
        DefaultInfo(
            files = depset(filtered_files),
        ),
    ]

filter_files_with_suffix = rule(
    implementation = _filter_files_with_suffix_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "suffix": attr.string(),
    },
)

def pyx_library(name, deps = [], cc_kwargs = {}, py_deps = [], srcs = [], **kwargs):
    """Compiles a group of .pyx / .pxd / .py files.

    First runs Cython to create .cpp files for each input .pyx or .py + .pxd
    pair. Then builds a shared object for each, passing "deps" and `**cc_kwargs`
    to each cc_binary rule (includes Python headers by default). Finally, creates
    a py_library rule with the shared objects and any pure Python "srcs", with py_deps
    as its dependencies; the shared objects can be imported like normal Python files.
    Added cc_kwargs on the basis of grpc rules.

    Args:
        name: Name for the rule.
        deps: C/C++ dependencies of the Cython (e.g. Numpy headers).
        cc_kwargs: cc_binary extra arguments such as copts, linkstatic, linkopts, features
        py_deps: Pure Python dependencies of the final library.
        srcs: .py, .pyx, or .pxd files to either compile or pass through.
        **kwargs: Extra keyword arguments passed to the py_library.
    """

    py_srcs = []
    pyx_srcs = []
    pxd_srcs = []
    for src in srcs:
        if src.endswith("__init__.py"):
            pxd_srcs.append(src)
        if src.endswith(".pyx") or (src.endswith(".py") and src[:-3] + ".pxd" in srcs):
            pyx_srcs.append(src)
        elif src.endswith(".py"):
            py_srcs.append(src)
        else:
            pxd_srcs.append(src)

    for filename in pyx_srcs:
        native.genrule(
            name = filename + "_cython_translation",
            srcs = [filename],
            outs = [filename.split(".")[0] + ".cpp"],
            cmd =
                "PYTHONHASHSEED=0 $(location @cython//:cython_binary) --cplus $(SRCS) --output-file $(OUTS)",
            tools = ["@cython//:cython_binary"] + pxd_srcs,
        )

    shared_objects = []
    for src in pyx_srcs:
        stem = src.split(".")[0]
        shared_object_name = stem + ".so"
        native.cc_binary(
            name = cc_kwargs.pop("name", shared_object_name),
            srcs = [stem + ".cpp"] + cc_kwargs.pop("srcs", []),
            deps = deps + ["@local_config_python//:python_headers"] + cc_kwargs.pop("deps", []),
            linkshared = cc_kwargs.pop("linkshared", 1),
            **cc_kwargs
        )
        shared_objects.append(shared_object_name)

    data = shared_objects[:]
    data += kwargs.pop("data", [])

    native.py_library(
        name = name,
        srcs = py_srcs,
        deps = py_deps,
        srcs_version = "PY2AND3",
        data = data,
        **kwargs
    )

def _cc_strip_impl(ctx):
    compilation_mode = ctx.var["COMPILATION_MODE"]
    if compilation_mode == "dbg":
        return [
            DefaultInfo(
                files = depset(ctx.files.srcs),
            ),
        ]
    cc_toolchain = find_cc_toolchain(ctx)
    input_files = ctx.files.srcs
    in_files_path = []
    output_files = []
    for s in ctx.attr.srcs:
        for f in s.files.to_list():
            in_files_path.append(f.path)
            output_files.append(ctx.actions.declare_file("%s_dir/%s" % (f.basename, f.basename), sibling = f))

    commands = [
        """chmod +w {obj} &&
        {obj_cpy} --only-keep-debug {obj} {dest} &&
        {obj_cpy} --add-gnu-debuglink={dest} {obj} &&
        {strip} --strip-all {obj} &&
        mkdir -p build/output/symbols && cp {dest} build/output/symbols
        mkdir -p {output_dir}
        cp -fr {obj} {output_dir}
        """.format(
            obj_cpy = cc_toolchain.objcopy_executable,
            strip = cc_toolchain.strip_executable,
            obj = src,
            dest = src + ".sym",
            output_dir = src + "_dir",
        )
        for src in in_files_path
    ]

    ctx.actions.run_shell(
        inputs = depset(
            direct = input_files,
            transitive = [
                cc_toolchain.all_files,
            ],
        ),
        outputs = output_files,
        progress_message = "CcStripping %s" % in_files_path,
        command = "".join(commands),
        mnemonic = "CcStrip",
    )
    return [
        DefaultInfo(
            files = depset(output_files),
        ),
    ]

cc_strip = rule(
    implementation = _cc_strip_impl,
    fragments = ["cpp"],
    attrs = {
        "_cc_toolchain": attr.label(
            default = Label(
                "@rules_cc//cc:current_cc_toolchain",  # copybara-use-repo-external-label
            ),
        ),
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "out": attr.output(),
    },
    toolchains = [
        "@rules_cc//cc:toolchain_type",  # copybara-use-repo-external-label
    ],
)
