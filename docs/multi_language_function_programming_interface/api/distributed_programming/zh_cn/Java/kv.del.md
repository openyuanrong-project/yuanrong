# YR.kv().del

包名：`package org.yuanrong.runtime.client`。

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

## YR.kv().del(key)

### 接口说明

#### public void del(String key) Throws YRException

提供类 Redis 的 del 接口，支持从数据系统删除已储存的二进制数据。

```java

YR.kv().del("key");
```

- 参数：

   - **key** (String) – 要删除的单个数据的键值。

- 抛出：

   - **YRException** - 数据系统删除失败，数据系统断连。

## YR.kv().del(keys)

### 接口说明

#### public List<String> del(List<String> keys) Throws YRException

提供类 Redis 的 del 接口，支持从数据系统删除已储存的二进制数据。

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<String> result = YR.kv().del(keys);
```

- 参数：

   - **key** (List<String>) – 要删除的多条数据的键值。

- 返回：

    List<String>，删除失败的键值。当键不存在时，认为处理成功。

- 抛出：

   - **YRException** - 数据系统删除失败，数据系统断连。
