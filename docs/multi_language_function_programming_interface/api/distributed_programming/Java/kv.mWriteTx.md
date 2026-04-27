# YR.kv().mWriteTx()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Invoking the interfaces in this chapter will trigger the initialization of the data system client. The runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential risk of Out-of-Memory (OOM) errors.

Therefore, when using the interfaces in this chapter, it is recommended to declare a larger memory resource specification for the functions that utilize these interfaces.

:::

## YR.kv().mWriteTx() Synchronous storage

:::{Note}

Constraint: When setting keys and values, users should avoid using special characters, as this may cause failures in storage or retrieval.

:::

### Interface description

#### public void mWriteTx(List\<String\> keys, List\<Object\> values, MSetParam mSetParam) Throws YRException

Provide the mWriteTx synchronous storage interface, supporting the serialization of a set of objects and saving them to the data system.

```java

MSetParam msetParam = new MSetParam();
List<String> keys = new ArrayList<String>(){{
    add("synchronous-key1");
}};
List<Object> vals = new ArrayList<Object>(){{
    add("synchronous-value1");
}};
YR.kv().mWriteTx(keys, vals, msetParam);
```

- Parameters:

   - **keys** (List\<String>\) – Assign a key to the stored data to identify it. Use this key to query the data.
   - **values** (List\<Object>\) – A set of objects that need to be cached.
   - **mSetParam** (MSetParam) – Configure attributes such as whether the object requires reliability.

- Throws:

   - **YRException** - The `mSetTx` operation failed, possibly due to disconnection from the data system, mismatched numbers of keys and values, empty keys, or the `ExistenceOpt` option being set to ``NONE``.

SetParam class description

| Field      | Type         | Description                                                         |
| --------- | ------------ | ------------------------------------------------------------ |
| writeMode | WriteMode    | Set the reliability of the data. When the server configuration supports secondary caching for reliability, such as Redis, this configuration can ensure data reliability. The default value is ``WriteMode.NONE_L2_CACHE``. |
| existence | ExistenceOpt | Indicates whether key overwriting is supported. The optional parameters are ``ExistenceOpt.NONE`` (supported, default parameter) and ``ExistenceOpt.NX`` (not supported, optional). |
| ttlSecond | int          | The data lifecycle, after which the key will be deleted. The default value is ``0``, indicating that the key will persist indefinitely until the `del` interface is explicitly called. |
| cacheType | CacheType    | Used to identify whether the allocated medium is memory or disk. The optional parameters are ``YR::CacheType::Memory`` (memory medium) and ``YR::CacheType::Disk`` (disk medium). The default value is ``YR::CacheType::Memory``. |

WriteMode enumeration type description

| Enumeration constants               | Description                                                     |
| ---------------------- | -------------------------------------------------------- |
| NONE_L2_CACHE          | Do not write to the second-level cache.                                         |
| WRITE_THROUGH_L2_CACHE | Synchronously write data to the second-level cache to ensure data reliability.                   |
| WRITE_BACK_L2_CACHE    | Asynchronously write data to the second-level cache to ensure data reliability.                   |
| NONE_L2_CACHE_EVICT    | Do not write to the second-level cache, and the data may be actively evicted by the system when system resources are insufficient. |
