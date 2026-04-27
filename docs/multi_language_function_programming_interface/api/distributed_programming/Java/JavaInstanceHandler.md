# JavaInstanceHandler

package: `org.yuanrong.call`.

## public class JavaInstanceHandler

Create an operation class for creating a Java stateful function instance.

:::{Note}

The JavaInstanceHandler class is the handle returned after creating a Java class instance; it is the return type of the [JavaInstanceCreator.invoke](JavaInstanceCreator.md) interface after creating a Java class instance.

Users can use the function method of JavaInstanceHandler to create a Java class instance member method handle and return the handle class JavaInstanceFunctionHandler.

:::

### Interface description

#### public JavaInstanceHandler(String instanceId, String functionId, String className)

The constructor of JavaInstanceHandler.

- Parameters:

   - **instanceId** - Java function instance ID.
   - **functionId** - Java function deployment returns an ID.
   - **className** - Java function class name.

#### public JavaInstanceHandler()

Default constructor of JavaInstanceHandler.

#### public void terminate(boolean isSync) throws YRException

The member method of the JavaInstanceHandler class is used to recycle cloud Java function instances.

It supports synchronous or asynchronous termination.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
javaInstanceHandler.terminate(true);
```

:::{Note}

When synchronous termination is not enabled, the default timeout for the current kill request is 30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous termination is enabled, the interface will block until the instance is completely exited.

:::

- Parameters:

   - **isSync** - Whether to enable synchronization. If true, it indicates sending a kill request with the signal quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the instance; if false, it indicates sending a kill request with the signal quantity of killInstance to the function-proxy, and the kernel asynchronously kills the instance.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void terminate() throws YRException

The member method of the JavaInstanceHandler class is used to recycle cloud Java function instances.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
javaInstanceHandler.terminate();
```

:::{Note}

The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception.

:::

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void clearHandlerInfo()

Both the user and the runtime java hold JavaInstanceHandler.

To ensure that JavaInstanceHandler cannot be used by the user after calling `Finalize`, the member variables of JavaInstanceHandler should be cleared.

#### public Map<String, String> exportHandler() throws YRException

The member method of the JavaInstanceHandler class allows users to obtain handle information, which can be serialized and stored in a database or other persistence tools.

When the tenant context is enabled, handle information can also be obtained through this method and used across tenants.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
Map<String, String> out = javaInstanceHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    Map<String, String> type, storing the handle information of JavaInstanceHandler through kv key-value pairs.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void importHandler(Map<String, String> input) throws YRException

The member method of the JavaInstanceHandler class allows users to import handle information, which can be obtained and deserialized from persistent tools such as databases.

When the tenant context is enabled, this method can also be used to import handle information from other tenant contexts.

```java

JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
javaInstanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- Parameters:

   - **input** - Store the handle information of JavaInstanceHandler through the kv key-value pair.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public &lt;R&gt; JavaInstanceFunctionHandler&lt;R&gt; function(JavaInstanceMethod&lt;R&gt; javaInstanceMethod)

The member method of the JavaInstanceHandler class is used to return the member function handle of the Java class instance on the cloud.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class))
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- Parameters:

   - **&lt;R&gt;** - the type of the object.
   - **javaInstanceMethod** - JavaInstanceMethod Class instance.

- Returns:

    JavaInstanceFunctionHandler Instance.
