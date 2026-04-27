# JavaInstanceHandler terminate

package: `org.yuanrong.call`.

## Interface description

### public void terminate() throws YRException

The member method of the [JavaInstanceHandler](JavaInstanceHandler.md) class is used to recycle cloud Java function instances.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
javaInstanceHandler.terminate();
```

:::{Note}

The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception.

:::

- Throws:

   - **YRException** - Unified exception types thrown.

### public void terminate(boolean isSync) throws YRException

The member method of the [JavaInstanceHandler](JavaInstanceHandler.md) class is used to recycle cloud Java function instances.

It supports synchronous or asynchronous termination.

:::{Note}

When synchronous termination is not enabled, the default timeout for the current kill request is 30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous termination is enabled, the interface will block until the instance is completely exited.

:::

- Parameters:

   - **isSync** - Whether to enable synchronization. If true, it indicates sending a kill request with the signal quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the instance; if false, it indicates sending a kill request with the signal quantity of killInstance to the function-proxy, and the kernel asynchronously kills the instance.

- Throws:

   - **YRException** - Unified exception types thrown.
