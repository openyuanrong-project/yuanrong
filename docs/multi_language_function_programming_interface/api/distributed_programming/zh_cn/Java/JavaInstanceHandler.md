# JavaInstanceHandler

包名：`org.yuanrong.call`。

## public class JavaInstanceHandler

创建 Java 有状态函数实例的操作类。

:::{Note}

类 JavaInstanceHandler 是创建 Java 类实例后返回的句柄。是接口 [JavaInstanceCreator.invoke](JavaInstanceCreator.md) 创建 Java 类实例后的返回值类型。

用户可以使用 JavaInstanceHandler 的 function 方法创建 Java 类实例成员方法句柄，并返回句柄类 [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md)。

:::

### Interface description

#### public JavaInstanceHandler(String instanceId, String functionId, String className)

JavaInstanceHandler 的构造函数。

- 参数：

   - **instanceId** - Java 函数实例 ID。
   - **functionId** - Java 函数部署返回的 ID。
   - **className** - Java 函数所属类名。

#### public JavaInstanceHandler()

JavaInstanceHandler 的默认构造函数。

#### public void terminate(boolean isSync) throws YRException

JavaInstanceHandler 类的成员方法，用于回收云上 Java 函数实例。

支持同步或异步 terminate。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
javaInstanceHandler.terminate(true);
```

:::{Note}

在不开启同步 terminate 时，当前 kill 请求的默认超时时间为 30 s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。在开启同步 terminate 时，该接口会阻塞等待，直至实例完全退出。

:::

- 参数：

   - **isSync** - 是否开启同步。若为 true，表示向 function-proxy 发送信号量为 killInstanceSync 的 kill 请求，内核同步 kill 实例；若为 false，表示向 function-proxy 发送信号量为 killInstance 的 kill 请求，内核异步 kill 实例。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void terminate() throws YRException

JavaInstanceHandler 类的成员方法，用于回收云上 Java 函数实例。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
javaInstanceHandler.terminate();
```

:::{Note}

当前 kill 请求的默认超时时间为 30 s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。

:::

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void clearHandlerInfo()

用户和 runtime java 都持有 javaInstancehandler。为了保证 javaInstancehandler 在调用 `Finalize` 后不能被用户使用，应该清除 javaInstancehandler 的成员变量。

#### public Map<String, String> exportHandler() throws YRException

JavaInstanceHandler 类的成员方法，用户通过该方法可获取句柄信息，可以将其序列化后存入数据库等持久化工具。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
Map<String, String> out = javaInstanceHandler.exportHandler();
// 序列化后存储到数据库或其他持久化工具中。
```

- 返回：

    Map<String, String>类型，通过 kv 键值对存放的 JavaInstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public void importHandler(Map<String, String> input) throws YRException

JavaInstanceHandler 类的成员方法，用户通过该方法可导入句柄信息，可以从数据库等持久化工具获取并反序列化导入信息。

```java

JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
javaInstanceHandler.importHandler(in);
// 输入参数通过从数据库或其他持久性存储中检索并反序列化获得。
```

- 参数：

   - **input** - 通过 kv 键值对存放的 JavaInstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public &lt;R&gt; JavaInstanceFunctionHandler&lt;R&gt; function(JavaInstanceMethod&lt;R&gt; javaInstanceMethod)

JavaInstanceHandler 类的成员方法，用于返回云上 Java 类实例的成员函数句柄。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class))
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **&lt;R&gt;** - 对象的类型。
   - **javaInstanceMethod** - JavaInstanceMethod 类实例。

- 返回：

    JavaInstanceFunctionHandler 实例。
