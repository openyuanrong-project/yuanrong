# YR.kv().read()

package: `package org.yuanrong.runtime.client`.

:::{Warning}

Invoking the interfaces in this chapter will trigger the initialization of the data system client. The runtime process will additionally consume 50MB of memory. When deploying with K8S, this may pose a potential risk of Out-of-Memory (OOM) errors.

Therefore, when using the interfaces in this chapter, it is recommended to declare a larger memory resource specification for the functions that utilize these interfaces.

:::

## YR.kv().read(key, timeoutSec, type)

### Interface description

#### public Object read(String key, int timeoutSec, Class&lt;?&gt; type) throws YRException

Supports retrieving cached data of specified types from the data system.

```java

String result = YR.kv().read("test", 10, String.class);
```

- Parameters:

   - **key** (String) – The key corresponding to the data being queried.
   - **timeoutSec** (int) – The timeout for the interface call, in seconds. The value must be greater than ``0`` or equal to ``-1``. A value of ``-1`` indicates no timeout restriction.
   - **type** (Class&lt;?&gt;) – Specify the type corresponding to data.

- Returns:

    Return the specified type of data queried.

- Throws:

   - **YRException** - The `read` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `write` operation was not successful before the `read` request was sent), or other similar issues.

## YR.kv().read(keys, timeoutSec, types, allowPartial)

### Interface description

#### public List&lt;Object&gt; read(List&lt;String&gt; keys, int timeoutSec, List&lt;Class&lt;?&gt;&gt; types, boolean allowPartial) throws YRException

Supports retrieving multiple cached data of specified types from the data system.

```java

List<String> keys = Arrays.asList("key1", "key2", "key3");
List<Class<?>> classes = Arrays.asList(String.class, String.class, String.class);
List<Object> results = YR.kv().read(keys, 10, classes, false);
```

- Parameters:

   - **keys** (List) – Specify the keys corresponding to multiple data items to query them in a single operation.
   - **timeoutSec** (int) – The timeout for the interface call, in seconds. The value must be greater than ``0`` or equal to ``-1``. A value of ``-1`` indicates no timeout restriction.
   - **types** (Class&lt;?&gt;) – Specify the types corresponding to multiple data.
   - **allowPartial** (boolean) – Whether partial return is allowed. When the value is ``true``, if only part of the query is successful, return the values that were successfully queried. Values that failed to be queried are filled with ``null``. When the value is ``false``, if any part of the query fails, an error is thrown.

- Returns:

    List&lt;Object&gt;, if all queries are successful, return the retrieved data.

- Throws:

   - **YRException** - The `read` operation failed, possibly due to disconnection from the data system, the data not being found in the data system (e.g., the `write` operation was not successful before the `read` request was sent), or other similar issues.
