# 提交作业

## 功能介绍

提交单个作业，可通过请求参数指定所属租户、亲和标签等信息。

## URI

POST `/api/jobs`

## 请求参数

请求 Header 参数

| **参数**     | **是否必选** | **参数类型** | **描述**                                                |
| ------------ | ------------ | ------------ | ------------------------------------------------------- |
| Content-Type | 是           | string       | 消息体类型。<br> **取值：** 建议填写 application/json。 |
| X-Tenant-Id  | 否           | string       | 作业所属租户。                                          |
| X-Pool-Label | 否           | string       | 作业实例所需亲和的资源池标签。                          |

<br> 请求 Body 参数

| **参数**             | **是否必选** | **参数类型**           | **描述**                                                                                                               |
| -------------------- | ------------ | ---------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| entrypoint           | 是           | string                 | 用于启动作业的入口命令。                                                                                               |
| metadata             | 否           | map[string]string      | 用户自定义的元信息，用于追踪或标记。                                                                                   |
| runtime_env          | 是           | map[string]interface{} | 作业运行环境配置，如依赖项、工作目录等。<br> 其中 **working_dir** 为必填项，可选填 env_vars 等其他参数。               |
| entrypoint_num_cpus  | 否           | float64                | 分配给作业的 CPU 数量，范围：[0.3,16]核。                                                                              |
| entrypoint_memory    | 否           | int                    | 分配给作业的 内存 数量，范围：[134217728,1125899906842624]字节。                                                       |
| entrypoint_num_gpus  | 否           | float64                | 分配给作业的 gpu 数量。                                                                                                |
| entrypoint_resources | 否           | map[string]float64     | 其他自定义资源配置，如 npu 等。                                                                                        |
| submission_id        | 否           | string                 | 作业实例 id，若请求未带，后端会自动生成一个，且前缀默认为 app- ，若没有后端会自动补齐。格式要求：`^[a-z0-9-]{1,64}$`。 |

<br> runtime_env 可选参数

| **参数**    | **是否必选** | **参数类型**      | **描述**                                                                                                                                                                                                                                                                                |
| ----------- | ------------ | ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| working_dir | 是           | string            | 用于启动作业的入口命令。<br> 有以下两种格式：<br> • 一种是传一个本地 zip 包文件所在路径：`file:///opt/openyuanrong/code/file.zip` ，后端会把该路径下的 zip 包文件解压到一个临时目录下后去引用；<br> • 另一种是传一个目录地址：`/opt/openyuanrong/code` ，后端会直接引用该路径下的代码包 |
| env_vars    | 否           | map[string]string | 用户自定义的元信息，用于追踪或标记。                                                                                                                                                                                                                                                    |

## 响应参数

| **参数**      | **参数类型** | **描述**  |
| ------------- | ------------ | --------- |
| submission_id | string       | 作业 id。 |

## 请求示例

POST {[frontend endpoint](api-frontend-endpoint)}/api/jobs

```json
{
  "entrypoint": "python script.py",
  "metadata": {
    "autoscenes_ids": "auto_1-test",
    "task_type": "task_1",
    "ttl": "1250"
  },
  "runtime_env": {
    "working_dir": "file:///usr1/deploy/file.zip",
    "env_vars": {
      "SOURCE_REGION": "suzhou_std",
      "DEPLOY_REGION": "suzhou_std"
    },
    "pip": ["numpy==1.24", "scipy==1.11.0"]
  },
  "entrypoint_num_cpus": 3,
  "entrypoint_memory": 134217728,
  "entrypoint_resources": {
    "NPU": 0
  },
  "submission_id": "app-1234"
}
```

## 响应示例

创建成功

```json
{
  "submission_id": "app-1234"
}
```

创建失败, 报错亲和节点不匹配

```text
"create app failed, submissionId:[app-xxxxx], err: \nthe instance cannot be scheduled within 18000 ms. which has been scheduled for 2 times. The latest failure: no available resource that meets the request requirements, The reasons are as follows:\n\t1 unit with [Affinity can't be Satisfied] requirements: [resource { aff { condition { subConditions { expressions { key: \"abc\" op { exits { } } } weight: 100 } orderPriority: true } } antiAff { } } ].\n"
```

## 错误码

请参见[错误码](error-code-rest-api)。
