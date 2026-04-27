# YR.kv().getWithParam()

包名：`package org.yuanrong.runtime.client`。

## YR.kv().getWithParam()

### 接口说明

#### public List<byte[]> getWithParam(List<String> keys, GetParams params) throws YRException

提供支持偏移读的接口 getWithParam, 支持偏移读取数据，默认超时时间为 5000 秒。

```java

String key = "kv-key";
String data = "kv-value";
SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
YR.kv().set(key, data.getBytes(StandardCharsets.UTF_8), setParam);
GetParam param = new GetParam.Builder().offset(0).size(2).build();
List<GetParam> paramList = new ArrayList<GetParam>(){{add(param);}};
GetParams params = new GetParams.Builder().getParams(paramList).build();
String result = new String(YR.kv().getWithParam(Arrays.asList("kv-key"), params).get(0));
```

- 参数：

   - **keys** (List<String>) – 为保存的数据设置一组键，用于标识这组数据。查询数据时使用其中的键进行查询。
   - **params** (GetParams) – 对应多个键读取时的 offset 和 size。

- 返回：

  List<byte[]>，返回查询到的一组二进制数据。获取任意一个 keys 对象成功，返回等长度的对象结果，失败的 key 对应下标的结果为空指针；获取 keys 全部失败时，抛出异常 4005，没有对象结果返回。

- 抛出：

   - **YRException** -

      - 集群模式运行出现错误会抛出异常，详见下表：

       | 异常错误码          | 描述         | 原因                                                |
       |----------------|------------|---------------------------------------------------|
       | 1001           | 入参错误       | Key 为空或含有无效字符， keys 和 params.getParams 的 size 不相等 |
       | 4005           | Get 操作错误   | Key 不存在，获取超时                                      |
       | 4201           | RocksDB 错误 | dsmaster 挂载盘有问题（磁盘满，有损坏等）                         |
       | 4202           | 共享内存限制     | 当前数据系统共享内存不足                                      |
       | 4203           | 数据系统操作磁盘失败 | 数据系统操作的目录没权限，或者其他问题                               |
       | 4204           | 磁盘空间满      | 磁盘空间满                                             |
       | 3002 | 内部通信错误     | 组件通信异常                                            |

      - 本地模式 key 超时，会抛出异常，错误码 4005。

#### public List<byte[]> getWithParam(List<String> keys, GetParams params, int timeoutSec) throws YRException

提供支持偏移读的接口 getWithParam, 支持偏移读取数据，可指定超时时间。

```java

String key = "kv-key";
String data = "kv-value";
SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
YR.kv().set(key, data.getBytes(StandardCharsets.UTF_8), setParam);
GetParam param = new GetParam.Builder().offset(0).size(2).build();
List<GetParam> paramList = new ArrayList<GetParam>(){{add(param);}};
GetParams params = new GetParams.Builder().getParams(paramList).build();
String result = new String(YR.kv().getWithParam(Arrays.asList("kv-key"), params, 10).get(0));
```

- 参数：

   - **keys** (List<String>) – 为保存的数据设置一组键，用于标识这组数据。查询数据时使用其中的键进行查询。
   - **params** (GetParams) – 对应多个键读取时的 offset 和 size。
   - **timeout** (int) – 单位为秒，取值范围[0, Integer.MAX_VALUE/1000)，为 -1 时表示永久阻塞等待。

- 返回：

    List<byte[]>，返回查询到的一组二进制数据。获取任意一个 keys 对象成功，返回等长度的对象结果，失败的 key 对应下标的结果为空指针；获取 keys 全部失败时，抛出异常 4005，没有对象结果返回。

- 抛出：

   - **YRException** -

      - 集群模式运行出现错误会抛出异常，详见下表：

       | 异常错误码          | 描述         | 原因                                                |
       |----------------|------------|---------------------------------------------------|
       | 1001           | 入参错误       | Key 为空或含有无效字符， keys 和 params.getParams 的 size 不相等 |
       | 4005           | Get 操作错误   | Key 不存在，获取超时                                      |
       | 4201           | RocksDB 错误 | dsmaster 挂载盘有问题（磁盘满，有损坏等）                         |
       | 4202           | 共享内存限制     | 当前数据系统共享内存不足                                      |
       | 4203           | 数据系统操作磁盘失败 | 数据系统操作的目录没权限，或者其他问题                               |
       | 4204           | 磁盘空间满      | 磁盘空间满                                             |
       | 3002 | 内部通信错误     | 组件通信异常                                            |

      - 本地模式 key 超时，会抛出异常，错误码 4005。

GetParams 类介绍

| 字段        | 类型             | 说明                         |
|-----------|----------------|----------------------------|
| getParams | List<GetParam> | GetParam 列表，对应多个键读取时的偏移参数。 |

GetParam 类介绍

| 字段     | 类型     | 说明        |
|--------|--------|-----------|
| offset | long   | 读取数据的偏移量。 |
| size   | long   | 读取数据的长度。  |
