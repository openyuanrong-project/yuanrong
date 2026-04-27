# YR.kv().mSetTx()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Calling the interfaces in this section will trigger the initialization of the data system client, and the runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential OOM risk. Therefore, when using the interfaces in this section, please declare a larger memory resource specification for the function using the interface.

:::

## YR.kv().mSetTx() Synchronous Storage

:::{Note}

Constraint: Users should avoid using special characters when setting keys and values, as this may cause storage or retrieval to fail.

:::

### Interface description

#### public void mSetTx(List<String> keys, List<byte[]> vals, MSetParam mSetParam) Throws YR Exception

The `mSetTx` interface is provided for synchronous storage, supporting the saving of a set of binary data to the data system.

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

- Parameters:

   - **keys** (List<String>) – A set of keys is assigned to the saved data to identify this group of data. When querying data, one of the keys is used for the query.
   - **values** (List<byte[]>) – A set of binary data that needs to be cached.
   - **setParam** (SetParam) – Configure attributes such as whether reliability is needed for the object.

- Throws:

   - **YRException** - It may be caused by data system disconnection, mismatched numbers of keys and vals, empty keys, or the ExistenceOpt option being set to NONE.

#### public void mSetTx(List<String> keys, List<byte[]> vals, List<Integer> lengths, MSetParam mSetParam) Throws YRException

The `set` interface is provided for synchronous storage, supporting the saving of binary data to the data system.

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

- Parameters:

   - **keys** (List<String>) – A set of keys is assigned to the saved data to identify this group of data. When querying data, one of the keys is used for the query.
   - **values** (List<byte[]>) – A set of binary data that needs to be cached.
   - **lengths** (List<Integer>) – The length of the binary data to be stored. The position of the length in the array should correspond to the position of the data in the vals array. The length of the list should be equal to the length of the keys. Users need to ensure the correctness of the len values themselves.
   - **setParam** (SetParam) – Configure attributes such as whether reliability is needed for the object.

- Throws:

   - **YRTaskException** - It may be caused by data system disconnection or the key containing illegal characters.

SetParam Class description

| Field     | Type         | Description                                                                                                                                                                                                                                   |
| --------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| writeMode | WriteMode    | Sets the reliability of the data. When the server configuration supports secondary caching to ensure reliability, such as with Redis service, this configuration can ensure data reliability. The default value is `WriteMode.NONE_L2_CACHE`. |
| existence | ExistenceOpt | Whether to support overwriting an existing Key. Optional parameters are `ExistenceOpt.NONE` (supported, default parameter) and `ExistenceOpt.NX` (unsupported, optional).                                                                     |
| ttlSecond | int          | The data lifecycle, after which it will be deleted; the default value is `0`, indicating that the key will exist indefinitely until explicitly deleted via the `del` interface.                                                               |
| cacheType | CacheType    | Identifies whether the allocated storage is memory-based or disk-based. Optional parameters are `YR::CacheType::Memory` (memory-based) and `YR::CacheType::Disk` (disk-based). The default value is `YR::CacheType::Memory`.                  |

WriteMode Enumeration Type description

| Enumeration Constant      | Description                                                                                                           |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| NONE\_L2\_CACHE           | Do not write to the second-level cache.                                                                               |
| WRITE\_THROUGH\_L2\_CACHE | Synchronously write data to the second-level cache to ensure data reliability.                                        |
| WRITE\_BACK\_L2\_CACHE    | Asynchronously write data to the second-level cache to ensure data reliability.                                       |
| NONE\_L2\_CACHE\_EVICT    | Do not write to the second-level cache, and when system resources are insufficient, it will be evicted by the system. |
