# YR.kv().get()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Invoking the interfaces in this chapter will trigger the initialization of the data system client. The runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential risk of Out-of-Memory (OOM) errors.

Therefore, when using the interfaces in this chapter, it is recommended to declare a larger memory resource specification for the functions that utilize these interfaces.

:::

## YR.kv().get(key)

### Interface description

#### public byte[] get(String key) Throws YRException

Provide a Redis-like `get` interface, supporting the retrieval of a single cached binary data item from the data system.

```java

byte[] result = YR.kv().get(key);
```

- Parameters:

   - **key** (String) – The key corresponding to the data being queried.

- Returns:

    Return the queried binary data as a byte[].

- Throws:

   - **YRException** - The `get` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `set` operation was not successful before the `get` request was sent), or other similar issues.

## YR.kv().get(keys)

### Interface description

#### public List<byte[]> get(List<String> keys) Throws YRException

Provide a Redis-like `get` interface, supporting the retrieval of multiple cached binary data items from the data system.

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<byte[]> results = YR.kv().get(keys);
```

- Parameters:

   - **key** (List) – Specify the keys corresponding to multiple data items to query them in a single operation.

- Returns:

    List<byte[]>, If all queries are successful, return the retrieved data.

- Throws:

   - **YRException** - The `get` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `set` operation was not successful before the `get` request was sent), or other similar issues.

## YR.kv().get(keys, allowPartial)

### Interface description

#### public List<byte[]> get(List<String> keys, boolean allowPartial) Throws YRException

Provide a Redis-like `get` interface, supporting the retrieval of multiple cached binary data items from the data system.

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<byte[]> results = YR.kv().get(keys, false);
```

- Parameters:

   - **key** (List) – Specify the keys corresponding to multiple data items to query them in a single operation.
   - **allowPartial** (boolean) – Whether partial return is allowed. When the value is ``true``, if only part of the query is successful, return the values that were successfully queried. Values that failed to be queried are filled with ``null``. When the value is ``false``, if any part of the query fails, an error is thrown.

- Returns:

    List<byte[]>, Return the data that was successfully queried. When partial success is allowed, the values corresponding to keys that failed to be queried are ``null``.

- Throws:

   - **YRException** - The `get` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `set` operation was not successful before the `get` request was sent), or other similar issues.

## YR.kv().get(keys, timeoutSec, allowPartial)

### Interface description

#### public List<byte[]> get(List<String> keys, int timeoutSec, boolean allowPartial) Throws YRException

Provide a Redis-like `get` interface, supporting the retrieval of multiple cached binary data items from the data system, with the ability to set a timeout for this call.

- Parameters:

   - **key** (List) – Specify the keys corresponding to multiple data items to query them in a single operation.
   - **timeoutSec** (int) – The timeout for the interface call, in seconds. The value must be greater than ``0`` or equal to ``-1``. A value of ``-1`` indicates no timeout restriction.
   - **allowPartial** (boolean) – Whether partial return is allowed. When the value is ``true``, if only part of the query is successful, return the values that were successfully queried. Values that failed to be queried are filled with ``null``. When the value is ``false``, if any part of the query fails, an error is thrown.

- Returns:

    List<byte[]>, Return the data that was successfully queried. When partial success is allowed, the values corresponding to keys that failed to be queried are ``null``.

- Throws:

   - **YRException** - The `get` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `set` operation was not successful before the `get` request was sent), or other similar issues.
