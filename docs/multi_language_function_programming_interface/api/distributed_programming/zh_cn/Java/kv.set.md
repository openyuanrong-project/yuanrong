# YR.kv().set()

包名：`package org.yuanrong.runtime.client`。

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

## YR.kv().set() 同步存储

:::{Note}

约束：用户在设置键与值时，应避免使用特殊字符，否则将会导致存储或获取失败。

:::

### 接口说明

#### public boolean set(String key, byte[] value)

提供 set 同步存储接口，支持保存二进制数据到数据系统。

```java

boolean result = YR.kv().set("synchronous-key-1", "val-1".getBytes(StandardCharsets.UTF_8));
```

- 参数：

   - **key** (String) – 为保存的数据设置一个键，用于标识该数据。查询数据时使用这个键进行查询。
   - **value** (byte[]) – 需要缓存的二进制数据。云外限制最大存储 ``100M``。

- 返回：

    bool，如果存储成功则返回 true，反之返回 false。

#### public void set(String key, byte[] value, SetParam setParam) Throws YRException

提供 set 同步存储接口，支持保存二进制数据到数据系统。

```java

SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
boolean result = YR.kv().set("synchronous-key-1", "val-1".getBytes(StandardCharsets.UTF_8), setParam);
```

- 参数：

   - **key** (String) – 为保存的数据设置一个键，用于标识该数据。查询数据时使用这个键进行查询。
   - **value** (byte[]) – 需要缓存的二进制数据。云外限制最大存储 ``100M``。
   - **setParam** (SetParam) – 为对象配置是否需要可靠性等属性。

- 抛出：

   - **YRException** - 可能由数据系统断连，key 含非法字符等导致。

SetParam 类介绍

| 字段      | 类型         | 说明                                                         |
| --------- | ------------ | ------------------------------------------------------------ |
| writeMode | WriteMode    | 设置数据的可靠性。当服务端配置支持二级缓存时用于保证可靠性时，比如 redis 服务，那么通过这个配置可以保证数据可靠性。默认值为 ``WriteMode.NONE_L2_CACHE``。 |
| existence | ExistenceOpt | 是否支持 Key 重复写入。可选参数为 ``ExistenceOpt.NONE``（支持，默认参数） 和 ``ExistenceOpt.NX``（不支持，可选）。 |
| ttlSecond | int          | 数据生命周期，超过会被删除；默认值为 ``0``，表示 key 会一直存在，直到显式调用 `del` 接口。 |
| cacheType | CacheType    | 用于标识分配的是内存介质还是磁盘介质。可选参数为 ``YR::CacheType::Memory``（内存介质） 和 ``YR::CacheType::Disk``（磁盘介质）。默认值为 ``YR::CacheType::Memory``。 |

WriteMode 枚举类型介绍

| 枚举常量               | 说明                                                     |
| ---------------------- | -------------------------------------------------------- |
| NONE_L2_CACHE          | 不写入二级缓存。                                         |
| WRITE_THROUGH_L2_CACHE | 同步写数据到二级缓存，保证数据可靠性。                   |
| WRITE_BACK_L2_CACHE    | 异步写数据到二级缓存，保证数据可靠性。                   |
| NONE_L2_CACHE_EVICT    | 不写入二级缓存，并且当系统资源不足时，会被系统主动驱逐。 |

## YR.kv().set() 异步存储

### 接口说明

#### public Future<Boolean> set(String key, byte[] value, KVCallback kvCallback)

提供 set 异步存储接口，支持保存二进制数据到数据系统。

```java

Future<Boolean> setFuture = YR.kv().set("asynchronous-key", "asynchronous-value".getBytes(StandardCharsets.UTF_8), callback);
```

:::{Note}

1、多个异步存储操作被同时调用时，不保证真正写入数据系统的顺序。如果对顺序有要求，请在下一次存储前确认上一次已成功或使用同步接口。

2、约束：用户在设置键与值时，应避免使用特殊字符，否则将会导致存储或获取失败。KVCallback 应该为非空。

:::

- 参数：

   - **key** (String) – 为保存的数据设置一个键，用于标识该数据。查询数据时使用这个键进行查询。
   - **value** (byte[]) – 需要缓存的二进制数据。
   - **kvCallback** (KVCallback) – 用户自定义 onComplete （详见 KVCallback）回调，异步 set 成功后自动触发。

- 返回：

    Future<Boolean>，如果存储成功，调用该 future 的 get 方法将会返回 true ，并会触发用户配置的 onComplete 回调，反之 get 方法则会返回 false 。
