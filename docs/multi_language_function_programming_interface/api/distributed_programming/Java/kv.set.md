# YR.kv().set()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Calling the interfaces in this section will trigger the initialization of the data system client, and the runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential OOM risk. Therefore, when using the interfaces in this section, please declare a larger memory resource specification for the function using the interface.

:::

## YR.kv().set() Synchronous storage

:::{Note}

Constraints: Users should avoid using special characters when setting keys and values, as this may cause storage or retrieval to fail.

:::

### Interface description

#### public boolean set(String key, byte[] value)

The `set` interface is provided for synchronous storage, supporting the saving of binary data to the data system.

```java

boolean result = YR.kv().set("synchronous-key-1", "val-1".getBytes(StandardCharsets.UTF_8));
```

- Parameters:

   - **key** (String) – A key is assigned to the saved data to identify it. When querying data, this key is used for the query.
   - **value** (byte[]) – The binary data that needs to be cached. The maximum storage limit outside the cloud is ``100M``.

- Returns:

    bool, Returns `true` if the storage is successful, otherwise returns `false`.

#### public void set(String key, byte[] value, SetParam setParam) Throws YRException

The `set` interface is provided for synchronous storage, supporting the saving of binary data to the data system.

```java

SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
boolean result = YR.kv().set("synchronous-key-1", "val-1".getBytes(StandardCharsets.UTF_8), setParam);
```

- Parameters:

   - **key** (String) – A key is assigned to the saved data to identify it. When querying data, this key is used for the query.
   - **value** (byte[]) – The binary data that needs to be cached. The maximum storage limit outside the cloud is ``100M``.
   - **setParam** (SetParam) – Configure attributes for the object, such as reliability.

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

## YR.kv().set() Asynchronous storage

### Interface description

#### public Future<Boolean> set(String key, byte[] value, KVCallback kvCallback)

The `set` interface is provided for asynchronous storage, supporting the saving of binary data to the data system.

```java

Future<Boolean> setFuture = YR.kv().set("asynchronous-key", "asynchronous-value".getBytes(StandardCharsets.UTF_8), callback);
```

:::{Note}

1. When multiple asynchronous storage operations are called simultaneously, the order in which data is actually written to the data system is not guaranteed. If the order is important, please ensure that the previous operation has succeeded before the next storage operation, or use the synchronous interface.

2. Constraints: Users should avoid using special characters when setting keys and values, as this may cause storage or retrieval to fail. The KVCallback should not be null.

:::

- Parameters:

   - **key** (String) – A key is set for the data to be saved to identify it. This key is used for querying the data.
   - **value** (byte[]) – The binary data that needs to be cached.
   - **kvCallback** (KVCallback) – User-defined `onComplete` callback (see KVCallback for details), which is automatically triggered upon successful asynchronous `set`.

- Returns:

    Future<Boolean>, If the storage is successful, calling the `get` method of this future will return `true` and will trigger the user-configured `onComplete` callback; otherwise, the `get` method will return `false`.
