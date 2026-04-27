# InstanceHandler

package: `org.yuanrong.call`.

## public class InstanceHandler

Create an operation class for creating a Java stateful function instance.

:::{Note}

The class InstanceHandler is the handle returned after creating a Java class instance; it is the return type of the interface [InstanceCreator.invoke](InstanceCreator.md) after creating a Java class instance.

Users can use the function method of the built-in InstanceHandler to create a Java class instance member method handle and return the handle class [InstanceFunctionHandler](InstanceFunctionHandler.md).

:::

### Interface description

#### public InstanceHandler(String instanceId, ApiType apiType)

The constructor of the InstanceHandler.

- Parameters:

   - **instanceId** - Java function instance ID.
   - **apiType** - The enumeration class has two values: Function and Posix.
     It is used internally by Yuanrong to distinguish function types. The default is Function.

#### public InstanceHandler()

Default constructor of InstanceHandler.

#### public VoidInstanceFunctionHandler function(YRFuncVoid0 func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 0 input parameters and 0 return values.

- Parameters:

   - **func** - YRFuncVoid0 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) instance.

#### public void terminate(boolean isSync) throws YRException

The member method of the InstanceHandler class is used to recycle cloud Java function instances.

It supports synchronous or asynchronous termination.

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate(true);
```

:::{Note}

When synchronous termination is not enabled, the default timeout for the current kill request is 30 seconds. In scenarios such as high disk load or etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception. When synchronous termination is enabled, the interface will block until the instance is completely exited.

:::

- Parameters:

   - **isSync** - Whether to enable synchronization. If true, it indicates sending a kill request with the signal quantity of killInstanceSync to the function-proxy, and the kernel synchronously kills the instance; if false, it indicates sending a kill request with the signal quantity of killInstance to the function-proxy, and the kernel asynchronously kills the instance.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void terminate() throws YRException

The member method of the InstanceHandler class is used to recycle cloud Java function instances.

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate();
```

:::{Note}

The default timeout for the current kill request is 30 seconds. In scenarios such as high disk load and etcd failure, the kill request processing time may exceed 30 seconds, causing the interface to throw a timeout exception. Since the kill request has a retry mechanism, users can choose not to handle or retry after capturing the timeout exception.

:::

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void clearHandlerInfo()

Both the user and the runtime java hold the instanceHandler.

To ensure that the instanceHandler cannot be used by the user after calling `Finalize`, the member variables of the instanceHandler should be cleared.

#### public Map<String, String> exportHandler() throws YRException

The member method of the InstanceHandler class allows users to obtain handle information, which can be serialized and stored in a database or other persistence tools.

When the tenant context is enabled, handle information can also be obtained through this method and used across tenants.

```java

InstanceHandler instHandler = YR.instance(MyYRApp::new).invoke();
Map<String, String> out = instHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    Map<String, String> type, The handle information of the InstanceHandler is stored in the kv key-value pair.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public void importHandler(Map<String, String> input) throws YRException

The member method of the InstanceHandler class allows users to import handle information, which can be obtained and deserialized from persistent tools such as databases.

When the tenant context is enabled, this method can also be used to import handle information from other tenant contexts.

```java

InstanceHandler instanceHandler = new InstanceHandler();
instanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- Parameters:

   - **input** - The handle information of the InstanceHandler is stored in the kv key-value pair.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public &lt;R&gt; InstanceFunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 0 input parameters and 1 return value.

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc0 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, R> InstanceFunctionHandler<R> function(YRFunc1<T0, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 1 input parameter and 1 return value.

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
ObjectRef ref = InstanceHandler.function(Counter::Add).invoke(5);
int res = (int)YR.get(ref, 100);
InstanceHandler.terminate();
```

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc1 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, R> InstanceFunctionHandler<R> function(YRFunc2<T0, T1, R> func)

Java functions in the cloud can call member functions of Java class instances in the cloud, supporting user functions with 2 input parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc2 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, R> InstanceFunctionHandler<R> function(YRFunc3<T0, T1, T2, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 3 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc3 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T3, R> InstanceFunctionHandler<R> function(YRFunc4<T0, T1, T2, T3, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 4 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc4 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T3, T4, R> InstanceFunctionHandler<R> function(YRFunc5<T0, T1, T2, T3, T4, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 5 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc5 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0> VoidInstanceFunctionHandler function(YRFuncVoid1<T0> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 1 input parameter and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **func** - YRFuncVoid1 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1> VoidInstanceFunctionHandler function(YRFuncVoid2<T0, T1> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 2 input parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **func** - YRFuncVoid2 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2> VoidInstanceFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 3 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **func** - YRFuncVoid3 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T4> VoidInstanceFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 4 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **func** - YRFuncVoid4 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T4, T5> VoidInstanceFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 5 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;T5&gt;** - Input parameter type.
   - **func** - YRFuncVoid5 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.
