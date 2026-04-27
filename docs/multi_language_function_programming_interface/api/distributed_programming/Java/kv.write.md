# YR.kv().Write()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Calling the interfaces in this section will trigger the initialization of the data system client, and the runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential OOM risk. Therefore, when using the interfaces in this section, please declare a larger memory resource specification for the function using the interface.

:::

## YR.kv().Write() Synchronous storage

:::{Note}

Constraints: Users should avoid using special characters when setting keys and values, as this may cause storage or retrieval to fail.

:::

### Interface description

#### public void Write(String key, Object value, ExistenceOpt existence) throws YRException

The `Write` interface is provided for synchronous storage, supporting the serialization and saving of `value` to the data system.

```java

YR.kv().Write("synchronous-key-1", "val-1", ExistenceOpt.NONE);
```

- Parameters:

   - **key** (String) – A key is assigned to the saved data to identify it. When querying data, this key is used for the query.
   - **value** (Object) – The object that needs to be cached.
   - **existence** (ExistenceOpt) – Whether to support overwriting an existing Key. Optional parameters are ``ExistenceOpt.NONE`` (supported, default parameter) and ``ExistenceOpt.NX`` (unsupported, optional).

- Throws:

   - **YRException** - It may be caused by data system disconnection or the key containing illegal characters.

#### public void Write(String key, Object value, SetParam setParam) throws YRException

The `Write` interface is provided for synchronous storage, supporting the serialization and saving of `value` to the data system.

```java

SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
YR.kv().Write("synchronous-key-1", "val-1", setParam);
```

- Parameters:

   - **key** (String) – A key is assigned to the saved data to identify it. When querying data, this key is used for the query.
   - **value** (Object) – The object that needs to be cached.
   - **setParam** (SetParam) – Configure attributes such as whether reliability is needed for the object.

- Throws:

   - **YRException** - It may be caused by data system disconnection or the key containing illegal characters.

SetParam class description

| Field      | Type         | Description                                                         |
| --------- | ------------ | ------------------------------------------------------------ |
| writeMode | WriteMode    | Set the reliability of the data. When the server configuration supports secondary caching for reliability, such as Redis, this configuration can ensure data reliability. The default value is ``WriteMode.NONE_L2_CACHE``. |
| existence | ExistenceOpt | Indicates whether key overwriting is supported. The optional parameters are `ExistenceOpt.NONE` (supported, default parameter) and `ExistenceOpt.NX` (not supported, optional). |
| ttlSecond | int          | The data lifecycle, after which the key will be deleted. The default value is ``0``, indicating that the key will persist indefinitely until the `del` interface is explicitly called. |
| cacheType | CacheType    | Used to identify whether the allocated medium is memory or disk. The optional parameters are `YR::CacheType::Memory` (memory medium) and `YR::CacheType::Disk` (disk medium). The default value is `YR::CacheType::Memory`. |

WriteMode enumeration type description

| Enumeration constants               | Description                                                     |
| ---------------------- | -------------------------------------------------------- |
| NONE_L2_CACHE          | Do not write to the second-level cache.                                         |
| WRITE_THROUGH_L2_CACHE | Synchronously write data to the second-level cache to ensure data reliability.                   |
| WRITE_BACK_L2_CACHE    | Asynchronously write data to the second-level cache to ensure data reliability.                   |
| NONE_L2_CACHE_EVICT    | Do not write to the second-level cache, and the data may be actively evicted by the system when system resources are insufficient. |
