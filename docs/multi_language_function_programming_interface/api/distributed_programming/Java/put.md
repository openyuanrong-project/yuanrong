# put

## YR.put()

:::{Warning}

Invoking the interfaces in this chapter will trigger the initialization of the data system client. The runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential risk of Out-of-Memory (OOM) errors.

Therefore, when using the interfaces in this chapter, it is recommended to declare a larger memory resource specification for the functions that utilize these interfaces.

:::

### Interface description

#### public static &lt;T&gt; ObjectRef put(T obj) throws YRException

Store the data object in the data system.

```java

ObjectRef<Integer> ref = YR.put(1);
```

- Parameters:

   - **&lt;T&gt;** - Input parameter type.
   - **obj** (Object) – The data objects to be stored must be contained within a List container (objects not stored in a List container are not allowed).

- Returns:

    ObjectRef: The unique identifier ("id") of the data within the data system. For more details, see the [ObjectRef](ObjectRef.md) documentation.

- Throws:

   - **YRException** - The passed-in `objectRef` must be stored within a List container; otherwise, the exception will be thrown.
   - **IOException** - The storage operation failed, which may be caused by disconnection from the data system or other related issues.

#### public static &lt;T&gt; ObjectRef put(T obj, CreateParam createParam) throws YRException

Store the data object in the data system.

```java

CreateParam createParam = new CreateParam();
createParam.setWriteMode(WriteMode.NONE_L2_CACHE);
createParam.setConsistencyType(ConsistencyType.PRAM);
createParam.setCacheType(CacheType.MEMORY);
ObjectRef<Integer> ref = YR.put(1, createParam);
```

- Parameters:

   - **&lt;T&gt;** - Input parameter type.
   - **obj** (Object) – The data objects to be stored must be contained within a List container (objects not stored in a List container are not allowed).
   - **createParam** (CreateParam) - Optional. Configure attributes for the object, such as whether reliability is required.

- Returns:

    ObjectRef: The unique identifier ("id") of the data within the data system. For more details, see the [ObjectRef](ObjectRef.md) documentation.

- Throws:

   - **YRException** - The passed-in `objectRef` must be stored within a List container; otherwise, the exception will be thrown.
   - **IOException** - The storage operation failed, which may be caused by disconnection from the data system or other related issues.

CreateParam 类介绍

| Field      | Type         | Description                            |
| --------- | ------------ | ------------------------------------------------------------ |
| writeMode | WriteMode    | Set the reliability of the data. When the server configuration supports secondary caching for reliability, such as Redis, this configuration can ensure data reliability. The default value is ``WriteMode.NONE_L2_CACHE``. |
| consistencyType | ConsistencyType | Data consistency configuration. In a distributed scenario, different levels of consistency semantics can be configured. Optional parameters include ``YR::ConsistencyType::PRAM`` (asynchronous) and ``YR::ConsistencyType::CAUSAL`` (causal consistency). The default value is ``YR::ConsistencyType::PRAM``. |
| cacheType | CacheType    | Used to identify whether the allocated medium is memory or disk. The optional parameters are `YR::CacheType::Memory` (memory medium) and `YR::CacheType::Disk` (disk medium). The default value is `YR::CacheType::Memory`. |
