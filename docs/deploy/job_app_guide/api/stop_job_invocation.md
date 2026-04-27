# 停止作业

## 功能介绍

当作业在 running 状态下，停止当前作业。

## URI

POST `/api/jobs/{submissionId}/stop`

路径参数：

| **参数**     | **是否必选** | **参数类型** | **描述**      |
| ------------ | ------------ | ------------ | ------------- |
| submissionId | 是           | String       | 作业实例 id。 |

## 响应参数

返回一个字符串，停止作业成功会返回 true ，停止失败会返回 false。

## 请求示例

POST {[frontend endpoint](api-frontend-endpoint)}/api/jobs/{submissionId}/stop

## 响应示例

停止成功

```text
true
```

停止失败

```text
false
```

## 错误码

请参见[错误码](error-code-rest-api)。
