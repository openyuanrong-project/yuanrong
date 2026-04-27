# YR.kv().mSetTx()

包名：`package org.yuanrong.runtime.client`。

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

## YR.kv().mSetTx() 同步存储

:::{Note}

约束：用户在设置键与值时，应避免使用特殊字符，否则将会导致存储或获取失败。

:::

### 接口说明

#### public void mSetTx(List<String> keys, List<byte[]> vals, MSetParam mSetParam) Throws YRException

提供 mSetTx 同步存储接口，支持保存一组二进制数据到数据系统。

```java

MSetParam msetParam = new MSetParam();
List<String> keys = new ArrayList<String>(){{
    add("synchronous-key1");
}};
List<byte[]> vals = new ArrayList<byte[]>(){{
    add("synchronous-value1".getBytes(StandardCharsets.UTF_8));
}};
YR.kv().mSetTx(keys, vals, msetParam);
```

- 参数：

   - **keys** (List<String>) – 为保存的数据设置一组键，用于标识这组数据。查询数据时使用其中的键进行查询。
   - **values** (List<byte[]>) – 需要缓存的一组二进制数据。
   - **setParam** (SetParam) – 为对象配置是否需要可靠性等属性。

- 抛出：

   - **YRException** - 可能由数据系统断连，keys 和 vals 数量不匹配，keys 为空，ExistenceOpt 选项为 NONE 等导致。

#### public void mSetTx(List<String> keys, List<byte[]> vals, List<Integer> lengths, MSetParam mSetParam) Throws YRException

提供 set 同步存储接口，支持保存二进制数据到数据系统。

```java

MSetParam msetParam = new MSetParam();
List<String> keys = new ArrayList<String>(){{
    add("synchronous-key1");
}};
List<byte[]> vals = new ArrayList<byte[]>(){{
    add("synchronous-value1".getBytes(StandardCharsets.UTF_8));
}};
List<Integer> lengths = new ArrayList<Integer>(){{
    add(18);
}};
YR.kv().mSetTx(keys, vals, lengths, msetParam);
```

- 参数：

   - **keys** (List<String>) – 为保存的数据设置一组键，用于标识这组数据。查询数据时使用其中的键进行查询。
   - **values** (List<byte[]>) – 需要缓存的一组二进制数据。
   - **lengths** (List<Integer>) – 需要存储的二进制数据的长度。数组内长度的位置需要和 vals 数组中数据的位置相对应。列表的长度要和 keys 相等。用户需要自行保证 len 值的正确性。
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
