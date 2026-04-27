# YR.kv().get()

包名：`package org.yuanrong.runtime.client`。

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

## YR.kv().get(key)

### 接口说明

#### public byte[] get(String key) Throws YRException

提供类 Redis 的 get 接口，支持从数据系统获取已缓存的单个二进制数据。

```java

byte[] result = YR.kv().get(key)
```

- 参数：

   - **key** (String) – 查询数据对应的 key。

- 返回：

    返回查询到的二进制数据，类型为 byte[]。

- 抛出：

   - **YRException** - get 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 set 成功就发送 get 请求）等。

## YR.kv().get(keys)

### 接口说明

#### public List<byte[]> get(List<String> keys) Throws YRException

提供类 Redis 的 get 接口，支持从数据系统获取多个已缓存的二进制数据。

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<byte[]> results = YR.kv().get(keys);
```

- 参数：

   - **key** (List) – 指定多个数据对应的键，一次查询多个数据。

- 返回：

    List<byte[]>，如果全部成功，则返回查询到的数据。

- 抛出：

   - **YRException** - get 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 set 成功就发送 get 请求）等。

## YR.kv().get(keys, allowPartial)

### 接口说明

#### public List<byte[]> get(List<String> keys, boolean allowPartial) Throws YRException

提供类 Redis 的 get 接口，支持从数据系统获取多个已缓存的二进制数据。

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<byte[]> results = YR.kv().get(keys, false);
```

- 参数：

   - **key** (List) – 指定多个数据对应的键，一次查询多个数据。
   - **allowPartial** (boolean) – 是否允许部分返回。当值为 ``true`` 时，如果查询只有部分成功，则返回查询成功的值。查询失败的值用 ``null`` 充填。当值为 ``false`` 时，如果查询中有任意一个失败则抛出错误。

- 返回：

    List<byte[]>，返回查询成功的数据。当允许部分成功时，查询失败的键对应的值为 ``null``。

- 抛出：

   - **YRException** - get 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 set 成功就发送 get 请求）等。

## YR.kv().get(keys, timeoutSec, allowPartial)

### 接口说明

#### public List<byte[]> get(List<String> keys, int timeoutSec, boolean allowPartial) Throws YRException

提供类 Redis 的 get 接口，支持从数据系统获取多个已缓存的二进制数据, 并设置本次调用的超时时间。

- 参数：

   - **key** (List) – 指定多个数据对应的键，一次查询多个数据。
   - **timeoutSec** (int) – 接口调用的超时时间，单位：秒。取值要求大于 ``0`` 或等于 ``-1``。``-1`` 表示不限制超时时间。
   - **allowPartial** (boolean) – 是否允许部分返回。当值为 ``true`` 时，如果查询只有部分成功，则返回查询成功的值。查询失败的值用 ``null`` 充填。当值为 ``false`` 时，如果查询中有任意一个失败则抛出错误。

- 返回：

    List<byte[]>，返回查询成功的数据。当允许部分成功时，查询失败的键对应的值为 ``null``。

- 抛出：

   - **YRException** - get 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 set 成功就发送 get 请求）等。
