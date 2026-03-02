# 查询作业详情

## 功能介绍

查询单个作业详情，获取驱动程序详细信息、作业状态等内容。

## URI

GET `/api/jobs/{submissionId}`

路径参数：

| **参数**     | **是否必选** | **参数类型** | **描述**      |
| ------------ | ------------ | ------------ | ------------- |
| submissionId | 是           | String       | 作业实例 id。 |

## 响应参数

| **参数**                  | **参数类型**                                        | **描述**                                         |
| ------------------------- | --------------------------------------------------- | ------------------------------------------------ |
| type                      | string                                              | 用于标识作业的类别。当前为：SUBMISSION。         |
| submission_id             | string                                              | 作业 id。                                        |
| driver_info               | [DriverInfo](api-data-struct-get-driverinfo) object | 与作业关联的驱动程序的详细信息。                 |
| status                    | string                                              | 作业当前的运行状态。                             |
| start_time                | string                                              | 作业的开始时间戳，单位为毫秒。                   |
| end_time                  | string                                              | 作业的结束时间戳，单位为毫秒。                   |
| metadata                  | map[string]string                                   | 作业的元数据。                                   |
| runtime_env               | map[string]interface{}                              | 作业的运行环境。                                 |
| driver_agent_http_address | string                                              | 暂未启用。                                       |
| driver_node_id            | string                                              | 驱动程序运行所在节点的 id。                      |
| driver_exit_code          | int32                                               | 驱动程序的退出码。                               |
| error_type                | string                                              | 类和识别问题由内部错误或用户提供的脚本错误引起。 |

(api-data-struct-get-driverinfo)=

<br> DriverInfo 类型参数

| **参数**        | **参数类型** | **描述**                     |
| --------------- | ------------ | ---------------------------- |
| id              | string       | 驱动程序的唯一标识符。       |
| node_ip_address | string       | 驱动程序所在节点的 IP 地址。 |
| pid             | string       | 驱动程序进程的进程号。       |

## 请求示例

GET {[frontend endpoint](api-frontend-endpoint)}/api/jobs/{submissionId}

## 响应示例

查询成功

```json
{
  "type": "SUBMISSION",
  "entrypoint": "python script_yr_sleep.py",
  "submission_id": "app-fe4ee1e5-d24f-4236-be15-7654e3c2229b",
  "driver_info": {
    "id": "app-fe4ee1e5-d24f-4236-be15-7654e3c2229b",
    "node_ip_address": "7.183.38.81",
    "pid": "748250"
  },
  "status": "SUCCEEDED",
  "start_time": "1755484077",
  "end_time": "1755484140",
  "metadata": {
    "autoscenes_ids": "auto_1-test",
    "task_type": "task_1",
    "ttl": "1250"
  },
  "runtime_env": {
    "envVars": "{\"DEPLOY_REGION\":\"suzhou_std\",\"SOURCE_REGION\":\"suzhou_std\"}",
    "pip": "pip3.9 install numpy==1.24 scipy==1.11.0 && pip3.9 check",
    "working_dir": "file:///home/disk/tk/file.zip"
  },
  "driver_agent_http_address": "",
  "driver_node_id": "node2-744947",
  "driver_exit_code": 0,
  "error_type": ""
}
```

查询失败, 报错 submissionId 找不到

```text
"not found app, submissionId: app-xxxxx, err: failed to get appjobInfo, submissionID: app-xxxxx"
```

## 错误码

请参见[错误码](error-code-rest-api)。
