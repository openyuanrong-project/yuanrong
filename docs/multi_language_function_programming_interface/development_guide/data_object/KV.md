# KV 缓存

openYuanrong 提供了近计算 KV 缓存能力，基于共享内存实现免拷贝的 KV 数据读写，实现高性能数据缓存，同时通过对接外部组件提供数据可靠性。

## 使用场景

openYuanrong  KV 接口使用近计算共享内存实现，所以数据读写时延较低，对于 MB 级以上的数据尤为明显。因此需要在函数间共享传递大块数据的场景，使用 KV 接口能取得较好的效果。

## 使用限制

- key 支持大写字母、小写字母、数字以及如下字符：`~.-/_!@#%^&*()+=:;`。
- key 不允许为空，支持的最大长度为 255 字节。
- value 的最大长度除 java 语言外没有限制，但不能超出 openYuanrong 数据系统配置的共享内存大小。Java 语言的 value 最大长度不能超过 `Integer.MAX_VALUE - 8`。
- 未写入二级缓存的数据，不保证数据可靠性，当发生故障时数据可能会丢失。

## 接口介绍

KV 接口支持传入二进制数据，也支持传入用户自定义对象，openYuanrong 自动做序列化/反序列化。数据写入时，支持设置以下参数：

- `writeMode`：用于设置数据可靠性，指定数据是否写入二级缓存，包含以下几种选项。
   - `NONE_L2_CACHE`：数据不写二级缓存，不保证可靠性，性能最优。当数据系统存储空间不足时，保留一份副本数据。
   - `WRITE_THROUGH_L2_CACHE`：数据同步写入二级缓存，可靠性最高，但性能较差，性能受限于二级缓存的性能。
   - `WRITE_BACK_L2_CACHE`：数据异步写入二级缓存，保证数据可靠性的同时，性能无明显下降。若数据异步写入二级缓存之前 openYuanrong 数据系统组件故障，可能会造成数据丢失。
   - `NONE_L2_CACHE_EVICT`：数据不写二级缓存，不保证可靠性。当数据系统空间不足时，会按 LRU 算法自动删除旧数据。
- `existence`：用于指定 key 是否允许重复写入。
- `ttlSecond`：用于指定数据的生命周期，当超出 TTL 自动删除。当配置为 0 表示数据长期有效，系统不会自动删除。

openYuanrong  KV 接口各语言支持 API 如下表所示。

| API 类型 | Python | Java | C++ |
| --------- | ---------------- | ------------- | -----------|
| 写入 | [yr.kv_write](../../api/distributed_programming/zh_cn/Python/yr.kv_write.rst)：写入二进制数据 | [YR.kv().set](../../api/distributed_programming/zh_cn/Java/kv.set.md)：写入二进制数据 | [YR::KV().Set](../../api/distributed_programming/zh_cn/Cpp/KV-Set.rst)：写入二进制数据<br>  [YR::KV().Write](../../api/distributed_programming/zh_cn/Cpp/KV-Write.rst)：写入用户自定义格式数据，openYuanrong 自动序列化<br>  [YR::KV().MSetTx](../../api/distributed_programming/zh_cn/Cpp/KV-MSetTx.rst)：批量写入数据多个二进制数据。多个数据具备事务性语义，保证同时成功或同时失败。<br>  [YR::KV().MWriteTx](../../api/distributed_programming/zh_cn/Cpp/KV-MWriteTx.rst)：批量写入多个用户自定义格式数据，openYuanrong 自动执行序列化。多个数据具备事务性语义，保证同时成功或同时失败。<br> |
| 读取 | [yr.kv_read](../../api/distributed_programming/zh_cn/Python/yr.kv_read.rst)：读取二进制数据 | [YR.kv().get](../../api/distributed_programming/zh_cn/Java/kv.get.md)：读取二进制数据 | [YR::KV().Get](../../api/distributed_programming/zh_cn/Cpp/KV-Get.rst)：读取数据，返回二进制数据。 <br> [YR::KV().Read](../../api/distributed_programming/zh_cn/Cpp/KV-Read.rst)：读取数据，将返回数据自动反序列化为用户自定义结构。该接口可直接将数据系统的共享内存反序列化为用户自定义结构，相比 YR::KV().Get 接口少了一次内存拷贝，性能更优。 <br> |
| 删除 | [yr.kv_del](../../api/distributed_programming/zh_cn/Python/yr.kv_del.rst) | [YR.kv().del](../../api/distributed_programming/zh_cn/Java/kv.del.md)  | [YR::KV().Del](../../api/distributed_programming/zh_cn/Cpp/KV-Del.rst) |

KV 接口的超时时间由环境变量 **DS_CONNECT_TIMEOUT_SEC** 指定，默认值为 1800 秒。

## 数据一致性

openYuanrong 的 KV 接口支持 Causal 级别数据读写一致性。一致性模型定义参见 [Consistency Models](https://jepsen.io/consistency/models){target="_blank"}。

## 数据溢出到磁盘

KV 数据存储在 openYuanrong 数据系统的共享内存中，当内存不足时，支持自动将数据溢出到磁盘并从内存中删除数据。当数据需要读取时，自动从磁盘中加载到共享内存。磁盘空间也不足时，如果数据已经写入到二级缓存，则自动将数据从本地磁盘和内存中删除。当数据需要读取时，自动从二级缓存加载到共享内存。

KV 数据溢出到磁盘的配置与数据对象相同，可参见[数据对象](./index.md)中的描述。

## 数据可靠性

openYuanrong KV 接口提供可靠性语义，在数据写入接口中通过 `writeMode` 参数配置数据可靠性级别。数据系统通过对接外部存储组件作为二级缓存实现数据可靠性，您需要在部署 openYuanrong 时配置二级缓存相关参数，当前支持的二级缓存组件有：OBS/SFS。

注意，若在部署时未配置二级缓存，使用数据写入接口，参数 `writeMode` 配置为 `WRITE_THROUGH_L2_CACHE` 或 `WRITE_BACK_L2_CACHE` 时会返回失败。

openYuanrong 集群部署时二级缓存参数配置参考如下。

```yaml
# 指定二级缓存的类型。可选值为：'obs', 'sfs', 'none'.
# 默认值为'none'，表示不支持二级缓存。
l2CacheType: "none"
```

对接各类外部组件的配置参数如下：

::::{tab-set}
:::{tab-item} obs

```yaml
obs:
  # The access key for obs AK/SK authentication. If the value of encryptKit is not plaintext, encryption is required.
  obsAccessKey: ""
  # The secret key for obs AK/SK authentication. If the value of encryptKit is not plaintext, encryption is required.
  obsSecretKey: ""
  # OBS endpoint. Example: "xxx.hwcloudtest.cn"
  obsEndpoint: ""
  # OBS bucket name.
  obsBucket: ""
  # Whether to enable the https in obs. false: use HTTP (default), true: use HTTPS
  obsHttpsEnabled: false
  # Use cloud service token rotation to connect obs.
  cloudServiceTokenRotation:
    # Whether to use ccms credential rotation mode to access OBS, default is false. If is enabled, need to specify
    # iamHostName, identityProvider, projectId, regionId at least.
    # In addition, obsEndpoint and obsBucket need to be specified.
    enable: false
    # Domain name of the IAM token to be obtained.
    iamHostName: ""
    # Provider that provides permissions for the ds-worker. Example: csms-datasystem.
    identityProvider: ""
    # Project id of the OBS to be accessed. Example: fb6a00ff7ae54a5fbb8ff855d0841d00.
    projectId: ""
    # Region id of the OBS to be accessed. Example: cn-north-7.
    regionId: ""
    # Whether to access OBS of other accounts by agency, default is false. If is true, need to specify tokenAgencyName
    # and tokenAgencyDomain.
    enableTokenByAgency: false
    # Agency name for proxy access to other accounts. Example: obs_access.
    tokenAgencyName: ""
    # Agency domain for proxy access to other accounts. Example: op_svc_cff.
    tokenAgencyDomain: ""
```

:::
:::{tab-item} sfs

```yaml
# The path to the mounted SFS.
sfsPath: ""
```

:::
::::
