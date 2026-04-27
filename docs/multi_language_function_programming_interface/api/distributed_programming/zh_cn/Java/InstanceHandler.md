# InstanceHandler

包名：`org.yuanrong.call`。

## public class InstanceHandler

创建 Java 有状态函数实例的操作类。

:::{Note}

类 InstanceHandler 是创建 Java 类实例后返回的句柄，是接口 [InstanceCreator.invoke](InstanceCreator.md) 创建 Java 类实例后的返回值类型。用户可以使用 InstanceHandler 内置的 function 方法创建 Java 类实例成员方法句柄，并返回句柄类 [InstanceFunctionHandler](InstanceFunctionHandler.md)。

:::

### 接口说明

#### public InstanceHandler(String instanceId, ApiType apiType)

InstanceHandler 的构造函数。

- 参数：

   - **instanceId** - Java 函数实例 ID。
   - **apiType** - 枚举类，包含两个值：``Function`` 和 ``Posix``。
     用于 Yuanrong 内部区分函数类型，默认 ``Function``。

#### public InstanceHandler()

InstanceHandler 的默认构造函数。

#### public VoidInstanceFunctionHandler function(YRFuncVoid0 func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 0 个入参，0 个返回值的用户函数。

- 参数：

   - **func** - YRFuncVoid0 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public void terminate(boolean isSync) throws YRException

InstanceHandler 类的成员方法，用于回收云上 Java 函数实例。支持同步或异步 terminate。

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate(true);
```

:::{Note}

在不开启同步 terminate 时，当前 kill 请求的默认超时时间为 30 秒，在磁盘高负载 etcd 故障等场景下，kill 请求处理时间可能超过 30 秒，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。在开启同步 terminate 时，该接口会阻塞等待，直至实例完全退出。

:::

- 参数：

   - **isSync** - 是否开启同步。若为 true，表示向 function-proxy 发送信号量 为 killInstanceSync 的 kill 请求，内核同步 kill 实例；若为 false，表示向 function-proxy 发送信号量为 killInstance 的 kill 请求，内核异步 kill 实例。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void terminate() throws YRException

InstanceHandler 类的成员方法，用于回收云上 Java 函数实例。

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate();
```

:::{Note}

当前终止请求的默认超时时间为 30 秒。在高磁盘负载和 etcd 故障等场景下，终止请求的处理时间可能会超过 30 秒，导致接口抛出超时异常。由于终止请求具备重试机制，用户可以选择在捕获超时异常后不进行处理，或者进行重试。

:::

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void clearHandlerInfo()

用户和 runtime java 都持有 instancehandler。为了保证 instancehandler 在调用 `Finalize` 后不能被用户使用，应该清除 instancehandler 的成员变量。

#### public Map<String, String> exportHandler() throws YRException

InstanceHandler 类的成员方法，用户通过该方法可获取句柄信息，可以将其序列化后存入数据库等持久化工具。

```java

InstanceHandler instHandler = YR.instance(MyYRApp::new).invoke();
Map<String, String> out = instHandler.exportHandler();
// 序列化后存储到数据库或其他持久化工具中。
```

- 返回：

    Map<String, String> 类型，通过 kv 键值对存放的 InstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void importHandler(Map<String, String> input) throws YRException

InstanceHandler 类的成员方法，用户通过该方法可导入句柄信息，可以从数据库等持久化工具获取并反序列化导入信息。

```java

InstanceHandler instanceHandler = new InstanceHandler();
instanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- 参数：

   - **input** - 通过 kv 键值对存放的 InstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public &lt;R&gt; InstanceFunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 0 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc0 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, R> InstanceFunctionHandler<R> function(YRFunc1<T0, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 1 个入参，1 个返回值的用户函数。

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
ObjectRef ref = InstanceHandler.function(Counter::Add).invoke(5);
int res = (int)YR.get(ref, 100);
InstanceHandler.terminate();
```

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc1 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, R> InstanceFunctionHandler<R> function(YRFunc2<T0, T1, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 2 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc2 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, R> InstanceFunctionHandler<R> function(YRFunc3<T0, T1, T2, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 3 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc3 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T3, R> InstanceFunctionHandler<R> function(YRFunc4<T0, T1, T2, T3, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 4 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc4 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T3, T4, R> InstanceFunctionHandler<R> function(YRFunc5<T0, T1, T2, T3, T4, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 5 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc5 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public \<T0> VoidInstanceFunctionHandler function(YRFuncVoid1\<T0> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 1 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **func** - YRFuncVoid1 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1> VoidInstanceFunctionHandler function(YRFuncVoid2<T0, T1> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 2 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **func** - YRFuncVoid2 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2> VoidInstanceFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 3 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **func** - YRFuncVoid3 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T4> VoidInstanceFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 4 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **func** - YRFuncVoid4 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T4, T5> VoidInstanceFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 5 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;T5&gt;** - 入参类型。
   - **func** - YRFuncVoid5 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。
