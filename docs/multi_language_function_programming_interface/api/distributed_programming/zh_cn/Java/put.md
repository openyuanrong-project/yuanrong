# put

## YR.put()

:::{Warning}

调用本章节的接口会触发数据系统客户端初始化，runtime 进程将会额外占用 50MB 内存，在 K8S 上部署时有潜在 OOM 风险。
因此，当使用本章节的接口时，请为使用接口的函数声明一个更大的内存资源规格。

:::

### 接口说明

#### public static &lt;T&gt; ObjectRef put(T obj) throws YRException

将数据对象存储在数据系统。

```java

ObjectRef<Integer> ref = YR.put(1);
```

- 参数：

   - **&lt;T&gt;** - 入参类型。
   - **obj** (Object) – 所需要存储的数据对象（不能为未经 List 容器储存的 objectRef）。

- 返回：

    ObjectRef。该数据在数据系统中的唯一标识 “ id ”，详见 [ObjectRef](ObjectRef.md) 介绍。

- 抛出：

   - **YRException** - 传入 objectRef 需由 List 容器储存，否则抛出该异常。
   - **IOException** - 储存失败，可能由数据系统断连等引起。

#### public static &lt;T&gt; ObjectRef put(T obj, CreateParam createParam) throws YRException

将数据对象存储在数据系统。

```java

CreateParam createParam = new CreateParam();
createParam.setWriteMode(WriteMode.NONE_L2_CACHE);
createParam.setConsistencyType(ConsistencyType.PRAM);
createParam.setCacheType(CacheType.MEMORY);
ObjectRef<Integer> ref = YR.put(1, createParam);
```

- 参数：

   - **&lt;T&gt;** - 入参类型。
   - **obj** (Object) – 所需要存储的数据对象（不能为未经 List 容器储存的 objectRef）。
   - **createParam** (CreateParam) - 可选。为对象配置是否需要可靠性等属性。

- 返回：

    ObjectRef。该数据在数据系统中的唯一标识 “ id ”，详见 [ObjectRef](ObjectRef.md) 介绍。

- 抛出：

   - **YRException** - 传入 objectRef 需由 List 容器储存，否则抛出该异常。
   - **IOException** - 储存失败，可能由数据系统断连等引起。

CreateParam 类介绍

| 字段      | 类型         | 说明                                                         |
| --------- | ------------ | ------------------------------------------------------------ |
| writeMode | WriteMode    | 设置数据的可靠性。当服务端配置支持二级缓存时用于保证可靠性时，比如 redis 服务，那么通过这个配置可以保证数据可靠性。默认值为 ``WriteMode.NONE_L2_CACHE``。 |
| consistencyType | ConsistencyType | 数据一致性配置。分布式场景可以配置不同级别的一致性语义。可选参数为 ``YR::ConsistencyType::PRAM``（异步）和 ``YR::ConsistencyType::PRAM``（因果一致性）。默认值为 ``YR::ConsistencyType::PRAM``。 |
| cacheType | CacheType    | 用于标识分配的是内存介质还是磁盘介质。可选参数为 ``YR::CacheType::Memory``（内存介质） 和 ``YR::CacheType::Disk``（磁盘介质）。默认值为 ``YR::CacheType::Memory``。 |
