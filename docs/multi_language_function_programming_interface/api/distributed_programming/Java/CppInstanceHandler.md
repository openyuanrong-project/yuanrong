# CppInstanceHandler

package: `org.yuanrong.call`.

## public class CppInstanceHandler

Create an operation class for creating a cpp stateful function instance.

:::{Note}

The CppInstanceHandler class is the handle returned after a Java function creates a cpp class instance; it is the return type of the [CppInstanceCreator.invoke](CppInstanceCreator.md) interface after creating a cpp class instance.

Users can use the function method of CppInstanceHandler to create a cpp class instance member method handle and return a handle class [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md).

:::

### Interface description

#### public CppInstanceHandler(String instanceId, String functionId, String className)

The constructor of CppInstanceHandler.

- Parameters:

   - **instanceId** - cpp function instance ID.
   - **functionId** - cpp function deployment returns ID.
   - **className** - cpp function class name.

#### public CppInstanceHandler()

Default constructor of CppInstanceHandler.

#### public void terminate() throws YRException

The member method of the CppInstanceHandler class is used to recycle the cloud-based cpp function instance.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
cppInstanceHandler.terminate();
```

:::{Note}

The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception.

:::

- Throws:

   - **YRException** - If the recycling instance fails, an YRException exception is thrown. For specific errors, see the error message description.

#### public void terminate(boolean isSync) throws YRException

The member method of the CppInstanceHandler class is used to recycle cloud-based cpp function instances.

It supports synchronous or asynchronous termination.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
cppInstanceHandler.terminate();
```

:::{Note}

When synchronous termination is not enabled, the default timeout for the current kill request is 30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous termination is enabled, this interface will block until the instance is completely exited.

:::

- Parameters:

   - **isSync** - Whether to enable synchronization. If true, it indicates sending a kill request with the signal quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the instance; if false, it indicates sending a kill request with the signal quantity of killInstance to the function-proxy, and the kernel asynchronously kills the instance.

- Throws:

   - **YRException** - If the recycling instance fails, an YRException exception is thrown. For specific errors, see the error message description.

#### public Map<String, String> exportHandler() throws YRException

Obtain instance handle information.

CppInstanceHandler class member method. Users can obtain handle information through this method, which can be serialized and stored in a database or other persistent tools.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
Map<String, String> out = cppInstanceHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    The handle information of CppInstanceHandler is stored in the kv key-value pair.

- Throws:

   - **YRException** - The export information failed and threw an YRException exception. For specific errors, see the error message description.

#### public &lt;R&gt; CppInstanceFunctionHandler&lt;R&gt; function(CppInstanceMethod&lt;R&gt; cppInstanceMethod)

The member method of the CppInstanceHandler class is used to return the handle of the member function of the cloud C++ class instance.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
CppInstanceFunctionHandler cppInstFuncHandler = cppInstanceHandler.function(
    CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInstFuncHandler.invoke(5);
int res = (int) YR.get(ref, 100);
```

- Parameters:

   - **&lt;R&gt;** - the type parameter.
   - **cppInstanceMethod** - CppInstanceMethod class instance.

- Returns:

    CppInstanceFunctionHandler Instance.
