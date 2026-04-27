def _impl(repository_ctx):
  grpc_repo_dir = repository_ctx.path(repository_ctx.attr.path).dirname
  upb_repo_dir = repository_ctx.path(str(grpc_repo_dir) + "/third_party/upb")
  repository_ctx.symlink(upb_repo_dir, "")


grpc_upb_repository = repository_rule(
    implementation=_impl,
    local=True,
    attrs={"path": attr.label(mandatory=True)})