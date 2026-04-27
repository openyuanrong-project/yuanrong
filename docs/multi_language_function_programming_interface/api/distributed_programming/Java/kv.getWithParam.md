# YR.kv().getWithParam()

package: `package org.yuanrong.runtime.client`.

## YR.kv().getWithParam()

### Interface description

#### public List<byte[]> getWithParam(List<String> keys, GetParams params) throws YRException

The interface getWithParam is provided, which supports offset reading of data with a default timeout of 5000 seconds.

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

- Parameters:

   - **keys** (List<String>) – A set of keys is assigned to the saved data to identify this group of data. When querying data, one of the keys is used for the query.
   - **params** (GetParams) – The offset and size corresponding to multiple key reads.

- Returns:

  List<byte[]>, Returns a set of binary data that has been queried. If the query for any one of the keys objects is successful, it returns an object result of the same length, with the result at the index corresponding to the failed key being a null pointer; if the query for all keys fails, it throws exception 4005 and does not return any object result.

- Throws:

   - **YRException** -

      - When running in cluster mode, if an error occurs, an exception will be thrown. For details, see the table below:

       | Exception error code          | Description        | Reason                                                |
       |----------------|------------|---------------------------------------------------|
       | 1001           | Invalid input parameters       | The key is empty or contains invalid characters, or the sizes of keys and params.getParams are not equal. |
       | 4005           | Get operation error   | The key does not exist, or the retrieval timed out.                                      |
       | 4201           | RocksDB Error | The dsmaster mount disk has problems (such as being full or corrupted).                         |
       | 4202           | Shared memory limit     | The current data system does not have enough shared memory.                          |
       | 4203           | Data system disk operation failed | The data system does not have permission to operate on the directory, or there are other issues.                      |
       | 4204           | Disk space is full      | Disk space is full                            |
       | 3002 | Internal communication error   | Component communication exception                             |

      - In local mode, if the key times out, an exception with error code 4005 will be thrown.

#### public List<byte[]> getWithParam(List<String> keys, GetParams params, int timeoutSec) throws YRException

The interface getWithParam is provided, which supports offset reading of data and allows specifying a timeout.

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

- Parameters:

   - **keys** (List<String>) – A set of keys is assigned to the saved data to identify this group of data. When querying data, one of the keys is used for the query.
   - **params** (GetParams) – The offset and size corresponding to multiple key reads.
   - **timeout** (int) – The unit is in seconds, with a value range of [0, Integer.MAX_VALUE/1000). A value of -1 indicates permanent blocking wait.

- Returns:

    List<byte[]>, Returns a set of binary data that has been queried. If the query for any one of the keys objects is successful, it returns an object result of the same length, with the result at the index corresponding to the failed key being a null pointer; if the query for all keys fails, it throws exception 4005 and does not return any object result.

- Throws:

   - **YRException** -

      - When running in cluster mode, if an error occurs, an exception will be thrown. For details, see the table below:

       | Exception error code          | Description        | Reason                                                |
       |----------------|------------|---------------------------------------------------|
       | 1001           | Invalid input parameters       | The key is empty or contains invalid characters, or the sizes of keys and params.getParams are not equal. |
       | 4005           | Get operation error   | The key does not exist, or the retrieval timed out.                                      |
       | 4201           | RocksDB Error | The dsmaster mount disk has problems (such as being full or corrupted).                         |
       | 4202           | Shared memory limit     | The current data system does not have enough shared memory.                          |
       | 4203           | Data system disk operation failed | The data system does not have permission to operate on the directory, or there are other issues.                      |
       | 4204           | Disk space is full      | Disk space is full                            |
       | 3002 | Internal communication error   | Component communication exception                             |

      - In local mode, if the key times out, an exception with error code 4005 will be thrown.

GetParams Class description

| Field     | Type           | Description                                                                              |
| --------- | -------------- | ---------------------------------------------------------------------------------------- |
| getParams | List<GetParam> | List of GetParam objects, corresponding to the offset parameters for multiple key reads. |

GetParam Class description

| Field | Type | Description |
|-------|------|-------------|
| offset | long | The offset from which to read the data. |
| size | long | The length of data to read. |
