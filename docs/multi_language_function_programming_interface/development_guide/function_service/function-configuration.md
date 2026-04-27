# 配置函数

本节向您介绍函数服务的可用配置项。

## 配置环境变量

通过配置环境变量，可以在不修改代码的前提下执行不同的代码片段。环境变量作为函数配置的一部分，以字符串键值对的方式存储，不同函数拥有独立的环境变量。

环境变量在 openYuanrong 系统中使用 AES256(Advanced Encryption Standard 256) 标准加密存储，在初始化函数实例时，会将环境变量解密后注入到函数实例上下文中。

环境变量的键必须以大小写字母开头，只能包含大小写字母、数字。

环境变量通过[注册函数](../../api/function_service/register_function.md)和[更新函数](../../api/function_service/update_function.md) API 配置，请求参数为 `environment`，示例如下。

```json
{
    "environment": {
        "key1": "value1"
    }
}
```

通过上下文 `context` 的 `getUserData` 接口获取环境变量。 

```python
context.getUserData('key1')
```

## 配置单实例并发度

并发度是 openYuanrong 分配给函数单个实例的最大线程数，即每个实例可以同时处理的请求个数。

默认情况下，函数实例在同一时间只能处理一个请求, 请求处理完再处理下一个请求。通过配置单实例并发度，可以让一个函数实例并发处理多个请求，减少处理相同数量的请求所需要的实例个数，在一定程度上降低函数冷启动时延。

单实例并发度通过[注册函数](../../api/function_service/register_function.md)和[更新函数](../../api/function_service/update_function.md) API 配置，请求参数为 `concurrentNum`，示例如下。

```json
{
    "concurrentNum": "2"
}
```

## 配置最大实例数

openYuanrong 会根据函数实例负载情况进行扩容，上限为配置的函数最大实例数。通过配置最大实例数，可以实现流量及函数最大使用资源量的控制。

最大实例数通过[注册函数](../../api/function_service/register_function.md)和[更新函数](../../api/function_service/update_function.md) API 配置，请求参数为 `maxInstance`，示例如下。

```json
{
    "maxInstance": "10"
}
```

## 配置预留实例数

预留实例是 openYuanrong 分配给函数的最小实例个数，是函数实例最少运行的数量。函数创建完成后，就会预先拉起对应数量的函数实例。openYuanrong 会在请求规模降低时对函数实例进行缩容，但是至少保持最小实例个数的实例处理用户请求。预留实例虽然会固定占用一部分资源，但减少了请求处理时启动实例的过程，可以降低时延。

预留实例数通过[注册函数](../../api/function_service/register_function.md)和[更新函数](../../api/function_service/update_function.md) API 配置，请求参数为 `minInstance`，示例如下。

```json
{
    "minInstance": "3"
}
```

## 配置预留实例标签

您可以为预留实例配置标签，在调用时指定请求调度到具有该标签的实例上。为不同使用场景配置不同数量的预留实例并打上标签，可实现更细粒度的请求调度策略。

相关 REST API：

- [创建预留实例配置](../../api/function_service/create_reserve_instance.md)。 
- [更新预留实例配置](../../api/function_service/update_reserve_instance.md)。
- [查询预留实例配置](../../api/function_service/get_reserve_instance.md)。
- [删除预留实例配置](../../api/function_service/delete_reserve_instance.md)。
