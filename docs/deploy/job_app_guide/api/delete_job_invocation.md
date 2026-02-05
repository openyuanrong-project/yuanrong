# 删除作业

## 功能介绍

当作业在最终态 Failed/Succeeded/Stopped 下，删除当前作业。

## URI

DELETE `/api/jobs/{submissionId}`

路径参数：

| **参数**     | **是否必选** | **参数类型** | **描述**      |
| ------------ | ------------ | ------------ | ------------- |
| submissionId | 是           | String       | 作业实例 id。 |

## 响应参数

返回一个字符串，删除作业成功会返回 true，删除失败会返回 false。

## 请求示例

DELETE {[frontend endpoint](api-frontend-endpoint)}/api/jobs/{submissionId}

## 响应示例

删除成功

```text
true
```

删除失败

```text
false
```

## 错误码

请参见[错误码](error-code-rest-api)。
