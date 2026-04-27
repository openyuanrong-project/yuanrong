# YR.kv().del

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Invoking the interfaces in this chapter will trigger the initialization of the data system client. The runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential risk of Out-of-Memory (OOM) errors.

Therefore, when using the interfaces in this chapter, it is recommended to declare a larger memory resource specification for the functions that utilize these interfaces.

:::

## YR.kv().del(key)

### Interface description

#### public void del(String key) Throws YRException

Provide a Redis-like `del` interface, supporting the deletion of stored binary data from the data system.

```java

YR.kv().del("key");
```

- Parameters:

   - **key** (String) – The key of the single data item to be deleted.

- Throws:

   - **YRException** - The data system failed to delete the data, possibly due to disconnection from the data system.

## YR.kv().del(keys)

### Interface description

#### public List<String> del(List<String> keys) Throws YRException

Provide a Redis-like `del` interface, supporting the deletion of stored binary data from the data system.

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<String> result = YR.kv().del(keys);
```

- Parameters:

   - **key** (List<String>) – The keys of the multiple data items to be deleted.

- Returns:

    List<String>, The keys that failed to be deleted. When a key does not exist, it is considered as successfully processed.

- Throws:

   - **YRException** - The data system failed to delete the data, possibly due to disconnection from the data system.
