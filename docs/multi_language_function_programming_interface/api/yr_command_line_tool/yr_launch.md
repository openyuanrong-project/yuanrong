# launch

仅启动单个组件，通常用于容器 entrypoint 场景。

## 用法

```shell
yr launch [OPTIONS] COMPONENT
```

## 参数

* `COMPONENT`：组件名称，例如 `etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`、`frontend`、`dashboard`、`collector`、`meta_service`、`function_scheduler`；也支持 `iam_server`、`function_manager`、`runtime_manager` 这类面向 K8s 部署的内部组件。
* `--inherit-env`：继承父进程环境变量（默认不继承）。
* `--env-subst KEY`：将配置文件中的 `{KEY}` 替换为环境变量 `KEY` 的值。可重复指定，或使用逗号分隔（如 `--env-subst A,B`）。

## 说明

* `launch` 只拉起一个组件，不会自动拉起依赖组件。
* 生产部署请优先使用 `yr start`，以获得完整的依赖编排、健康检查与会话管理能力。

## 样例

```shell
yr launch etcd
```

```shell
yr launch --inherit-env --env-subst POD_IP,HOST_IP function_proxy
```
