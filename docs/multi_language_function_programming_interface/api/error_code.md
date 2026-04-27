# 错误码

## SDK 错误码

```{csv-table}
:header: >
:    "错误码", "描述", "解决方法"
:widths: 15, 20, 40
:align: left

1001, "参数校验失败。", "根据 message 提示，恢复异常参数，重新调用函数。"
1002, "资源不足。", "检查是否申请资源超过内核所管理的资源。"
1003, "找不到实例。", "检查用例是否被正常申请扩容。"
1004, "重复请求。", "联系我们。"
1005, "请求已达上限。", "间隔一定时间后重试。"
1006, "资源错误无法调度。", "检查配置资源是否有误，申请的资源是否超过每个数据面所管理的资源或者超过允许的范围。"
1007, "执行了 exit 或 kill 15 后实例退出。", "检查失败的 runtime 失败原因并进行修复。"
1008, "ExtensionMeta 配置错误。", "查看 ExtensionMeta 配置，并根据提示信息修改。"
1010, "成组实例创建失败。", "查看成组实例创建参数 InvokeOptions 配置，查看是否配置问题。"
1011, "成组实例退出失败。", "查看成组实例创建参数 InvokeOptions 配置，查看是否配置问题。"
1012, "local 限流实例创建失败。", "检查 local scheduler 创建流量限制。"
1013, "实例被驱逐。", "检查失败的 runtime 失败原因并进行修复。"
1014, "鉴权失败。", "检查 IAM 配置检验策略。"
1015, "函数元信息不存在。", "检查是否已部署函数。"
1016, "实例信息不正确。", "联系我们。"
1017, "实例调度被取消。", "create 请求调度时被取消，可能的情况是在调度超时时间内没有可用的资源、执行过程中调用 Finalize 等。"
2001, "用户代码无法加载。", "检查用户 so 是否正常、权限是否配置正确等。"
2002, "用户代码出现 core dump 或者异常。", "根据相关堆栈和 msg，检查用户函数错误并修改。"
3001, "连接 runtime 失败。", "联系我们。"
3002, "通讯错误。", "查看环境网络是否正常。"
3003, "内部错误。", "联系我们。"
3004, "frontend 连不上 busproxy。", "检查 frontend 所在节点的 busproxy 是否正常运行。"
3005, "etcd 存取异常。", "收集环境信息，联系我们。"
3006, "busproxy 断链。", "收集环境信息，联系我们。"
3007, "redis 存取异常。", "收集环境信息，联系我们。"
3008, "K8S 不可达。", "收集环境信息，联系我们。"
3009, "function agent 执行失败。", "收集环境信息，联系我们。"
3010, "状态机执行失败。", "收集环境信息，联系我们。"
3011, "local scheduler 执行失败。", "收集环境信息，联系我们。"
3012, "runtime manager 执行失败。", "收集环境信息，联系我们。"
3013, "instance manager 执行失败。", "收集环境信息，联系我们。"
3014, "local scheduler 异常。", "收集环境信息，联系我们。"
4001, "init 使用有误。", "修改 init 的使用方式。"
4002, "init 连接失败。", "查看 init 连接地址是否正确，环境网络是否正常。"
4003, "反序列失败。", "联系我们。"
4004, "本地模式所调用的实例不存在。", "检查实例是否创建正确。"
4005, "get 获取失败。", "确定 get 超时时间是否合理，如确定无问题则联系我们。"
4006, "函数使用方式有误。", "根据提示 msg 修改调用函数的使用方式。"
4007, "create 错误。", "根据提示 msg 修改相关 create 参数。"
4008, "invoke 错误。", "根据提示 msg 修改相关 invoke 参数。"
4009, "kill 错误。", "根据提示 msg 修改相关 kill 参数。"
4201, "操作 rockdb 失败。", "收集环境信息，联系我们。"
4202, "读写的数据超出了共享内存的大小。", "查看配置共享内存是否符合预期。"
4203, "操作磁盘文件失败。", "收集环境信息，联系我们。"
4204, "磁盘空间不足。", "收集环境信息，联系我们。"
4205, "数据系统连接失败。", "收集环境信息，联系我们。"
4299, "其他数据系统错误。", "收集环境信息，联系我们。"
4306, "依赖的请求失败。", "根据提示 msg 修改相关 create 参数。"
9000, "执行了 finalize 后实例退出。", "查看代码是否在使用 Finalize 或者 Terminate 接口后又发起调用请求。"
```

:::{note}

以下类型异常可稍后重试，如无法解决，联系我们。

- Internal Server error。
- timeout 类异常。

:::

(error-code-rest-api)=

## REST API 错误码 

| **错误码** | **错误信息**                     | **描述**                                     |
| ---------- | -------------------------------- | -------------------------------------------- |
| 4037       | session config invalid           | 用户指定实例会话不合法。                     |
| 4101       | FunctionNameExist                | 函数名已存在。                               |
| 4102       | AliasNameAlreadyExists           | 别名已存在。                                 |
| 4103       | TotalRoutingWeightNotOneHundred  | 总路由权重不是 100。                         |
| 4104       | InvalidUserParam                 | 用户输入参数无效。                           |
| 4105       | FunctionVersionDeletionForbidden | 函数版本有别名，因此不能删除。               |
| 4106       | LayerVersionSizeOutOfLimit       | 层版本超出限制。                             |
| 4107       | TenantLayerSizeOutOfLimit        | 租户层大小超出限制。                         |
| 4108       | LayerVersionNumOutOfLimit        | 层版本号超出限制。                           |
| 4109       | TriggerNumOutOfLimit             | 函数触发器数量超出限制。                     |
| 4110       | FunctionVersionOutOfLimit        | 函数版本超过限制。                           |
| 4111       | AliasOutOfLimit                  | 函数别名数超过限制。                         |
| 4112       | LayerIsUsed                      | 如果函数使用该层，则不能删除。               |
| 4113       | LayerVersionNotFound             | 层版本未创建。                               |
| 4114       | RepeatedPublishmentError         | 两次发布没有区别。                           |
| 4115       | FunctionNotFound                 | 函数未创建或已删除。                         |
| 4116       | FunctionVersionNotFound          | 函数版本未创建或已删除。                     |
| 4117       | AliasNameNotFound                | 别名未创建或已删除。                         |
| 4118       | TriggerNotFound                  | 触发器未创建或已删除。                       |
| 4119       | LayerNotFound                    | 层未创建或已删除。                           |
| 4120       | PoolNotFound                     | 池未创建或已删除。                           |
| 4123       | PageInfoError                    | 页面信息超出列表边界。                       |
| 4124       | TriggerPathRepeated              | 触发器的路径被其他触发器使用。               |
| 4125       | DuplicateCompatibleRuntimes      | 在兼容运行时列表中发现重复项。               |
| 4126       | CompatibleRuntimeError           | 一些兼容的运行时不在配置列表中。             |
| 4127       | ZipFileCountError                | zip 文件中的文件数量超出配置限制。           |
| 4128       | ZipFilePathError                 | zip 文件中的某些文件路径无效。               |
| 4129       | ZipFileUnzipSizeError            | 解压缩文件大小超过配置限制。                 |
| 4130       | ZipFileSizeError                 | 压缩包大小超出配置限制。                     |
| 4134       | RevisionIDError                  | 请求的修订 ID 与存储中操作的条目 ID 不匹配。 |
| 121016     | SaveFileError                    | 处理包的 zip 文件时，保存临时文件时出错。    |
| 121017     | UploadFileError                  | 上传文件时处理包 zip 文件出错。              |
| 121018     | EmptyAliasAndVersion             | 别名名称和版本为空。                         |
| 121019     | ReadingPackageTimeout            | 读取软件包时超时。                           |
| 121026     | BucketNotFound                   | 未找到特定业务 ID 的桶信息。                 |
| 121029     | ZipFileError                     | 处理包的 zip 文件时出错。                    |
| 121030     | DownloadFileError                | 从包存储中获取预签名下载 URL 时出错。        |
| 121032     | DeleteFileError                  | 删除包存储对象时出错。                       |
| 121036     | InvalidFunctionLayer             | 该层的租户信息与其请求的上下文不匹配。       |
| 121046     | InvalidQueryURN                  | 查询请求中的 URN 无效。                      |
| 121047     | InvalidQualifier                 | 请求中的限定符无效，它表示版本或别名。       |
| 121048     | ReadBodyError                    | 处理 HTTP 请求的正文时发生错误。             |
| 121049     | AuthCheckError                   | 当身份验证 HTTP 请求时发生错误。             |
| 121052     | InvalidJSONBody                  | 解析 HTTP 请求的正文为 JSON 格式时出错。     |
| 121057     | TriggerIDNotFound                | 请求中的触发器 ID 在存储中找不到。           |
| 121058     | FunctionNameFormatErr            | 存储中函数名称格式错误。                     |
| 122001     | KVNotFound                       | etcd 没有这样的 kv。                         |
| 122002     | EtcdError                        | etcd 中的错误。                              |
| 122003     | TransactionFailed                | etcd 事务出错。                              |
| 122004     | UnmarshalFailed                  | 反序列化为 JSON 时出错。                     |
| 122005     | MarshalFailed                    | 序列化为字符串时出错。                       |
| 122006     | VersionOrAliasEmpty              | 版本或别名是空字符串。                       |
| 122007     | ResourceIDEmpty                  | 资源 ID 为空字符串。                         |
| 122008     | NoTenantInfo                     | 租户信息不存在。                             |
| 122009     | NoResourceInfo                   | 资源信息不存在。                             |
| 130600     | InvalidParamErrorCode            | 解析 HTTP 请求的参数时出错。                 |
| 150424     | function metadata not found      | 找不到函数元数据。                           |
| 150500     | internal system error                      | 系统内部错误。                               |
| 150444     | instance label not found         | 指定实例 label 不存在。                      |
| 330404     | subcribe stream failed             | 订阅流失败。                                 |
