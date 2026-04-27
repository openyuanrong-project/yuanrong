def _impl(repository_ctx):
    path = repository_ctx.attr.path
    if path[0] != "/":
        cur_dir = repository_ctx.path(Label("//:BUILD.bazel")).dirname
        path = repository_ctx.path(str(cur_dir) + "/" + path)
    result = repository_ctx.execute(["cp", "-fr", str(path) + "/.", "."])
    if result.return_code != 0:
        fail("Failed to cp (%s): %s, %s" % (result.return_code, result.stderr, result.stdout))
    for patch_file in repository_ctx.attr.patch_files:
        patch_file = str(repository_ctx.path(patch_file).realpath)
        result = repository_ctx.execute(["patch", "-N", "-p0", "-i", patch_file])
        if result.return_code != 0:
            fail("Failed to patch (%s): %s, %s" % (result.return_code, result.stderr, result.stdout))
    if repository_ctx.attr.build_file:
       repository_ctx.symlink(repository_ctx.attr.build_file, repository_ctx.path("BUILD"))

local_patched_repository = repository_rule(
    implementation=_impl,
    attrs={
        "path": attr.string(mandatory=True),
        "patch_files": attr.label_list(allow_files=True),
        "build_file": attr.label(allow_single_file=True),
    },
    local = True)