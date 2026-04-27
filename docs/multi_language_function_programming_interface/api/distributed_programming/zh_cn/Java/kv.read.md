# YR.kv().read()

包名：`package org.yuanrong.runtime.client`。

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

## YR.kv().read(key, timeoutSec, type)

### 接口说明

#### public Object read(String key, int timeoutSec, Class&lt;?&gt; type) throws YRException

支持从数据系统获取已缓存的指定类型数据。

```java

String result = YR.kv().read("test", 10, String.class);
```

- 参数：

   - **key** (String) – 查询数据对应的 key。
   - **timeoutSec** (int) – 接口调用的超时时间，单位：秒。取值要求大于 ``0`` 或等于 ``-1``。``-1`` 表示不限制超时时间。
   - **type** (Class&lt;?&gt;) – 指定查询数据的类型。
- 返回：

    返回查询到的指定类型数据。

- 抛出：

   - **YRException** - read 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 write 成功就发送 read 请求）或反序列化为指定类型失败等。

## YR.kv().read(keys, timeoutSec, types, allowPartial)

### 接口说明

#### public List&lt;Object&gt; read(List&lt;String&gt; keys, int timeoutSec, List&lt;Class&lt;?&gt;&gt; types, boolean allowPartial) throws YRException

支持从数据系统获取多个已缓存的指定类型数据。

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<Class<?>> classes = Arrays.asList(String.class, String.class, String.class);
List<Object> results = YR.kv().read(keys, 10, classes, false);
```

- 参数：

   - **keys** (List) – 指定多个数据对应的键，一次查询多个数据。
   - **timeoutSec** (int) – 接口调用的超时时间，单位：秒。取值要求大于 ``0`` 或等于 ``-1``。``-1`` 表示不限制超时时间。
   - **types** (List&lt;Class&lt;?&gt;&gt;) – 指定多个数据对应的类型。
   - **allowPartial** (boolean) – 是否允许部分返回。当值为 ``true`` 时，如果查询只有部分成功，则返回查询成功的值。查询失败的值用 ``null`` 充填。当值为 ``false`` 时，如果查询中有任意一个失败则抛出错误。

- 返回：

    List&lt;Object&gt;，返回查询成功的数据。当允许部分成功时，查询失败的键对应的值为 ``null``。

- 抛出：

   - **YRException** - read 失败，可能是由于数据系统断连，数据系统内找不到该数据（还未 write 成功就发送 read 请求）或反序列化为指定类型失败等。
