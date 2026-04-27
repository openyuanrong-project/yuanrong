# 删除会话

## 功能介绍

该 API 用于删除 AI Agent 会话在数据系统中的持久化数据。

## URI

`DELETE /serverless/v1/functions/{function-urn}/sessions/{sessionId}`

| **参数** | **是否必选** | **参数类型** | **描述** |
| :--- | :--- | :--- | :--- |
| `function-urn` | 是 | string | 函数的版本 URN。 |
| `sessionId` | 是 | string | 会话 ID。 |

## 请求参数

### 请求 Header 参数

| **参数** | **是否必选** | **参数类型** | **描述** |
| :--- | :--- | :--- | :--- |
| Authorization | 是 | string | 鉴权信息。 |

## 响应参数

| **名称** | **类型** | **描述** |
| :--- | :--- | :--- |
| `code` | int | 返回码。仅在失败时返回。 |
| `message` | String | 返回信息。仅在失败时返回。 |

## 响应示例

### 正常响应

- **HTTP 状态码**: ``200``
- **Body**: ``(空)``

### 错误响应

```json
{
    "code": 200400,
    "message": "sessionId is empty"
}
```

或

```json
{
    "code": 200404,
    "message": "session not found"
}
```
