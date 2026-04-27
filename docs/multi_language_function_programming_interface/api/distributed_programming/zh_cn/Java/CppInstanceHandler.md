# CppInstanceHandler

包名：`org.yuanrong.call`。

## public class CppInstanceHandler

创建 cpp 有状态函数实例的操作类。

:::{Note}

类 CppInstanceHandler 是 Java 函数创建 cpp 类实例后返回的句柄。是接口 [CppInstanceCreator.invoke](CppInstanceCreator.md) 创建 cpp 类实例后的返回值类型。

用户可以使用 CppInstanceHandler 的 function 方法创建 cpp 类实例成员方法句柄，并返回句柄类 CppInstanceFunctionHandler。

:::

### 接口说明

#### public CppInstanceHandler(String instanceId, String functionId, String className)

CppInstanceHandler 的构造函数。

- 参数：

   - **instanceId** - cpp 函数实例 ID。
   - **functionId** - cpp 函数部署返回的 ID。
   - **className** - cpp 函数所属类名。

#### public CppInstanceHandler()

CppInstanceHandler 的默认构造函数。

#### public void terminate() throws YRException

CppInstanceHandler 类的成员方法，用于回收云上 cpp 函数实例。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
cppInstanceHandler.terminate();
```

:::{Note}

当前 kill 请求的默认超时时间为 30s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。

:::

- 抛出：

   - **YRException** - 回收实例失败抛出 `YRException` 异常，具体错误详见错误信息描述。

#### public void terminate(boolean isSync) throws YRException

CppInstanceHandler 类的成员方法，用于回收云上 cpp 函数实例。

支持同步或异步 terminate。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
cppInstanceHandler.terminate();
```

:::{Note}

在不开启同步 terminate 时，当前 kill 请求的默认超时时间为 30s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。在开启同步 terminate 时，该接口会阻塞等待，直至实例完全退出。

:::

- 参数：

   - **isSync** - 是否开启同步。若为 true，表示向 function-proxy 发送信号量为 killInstanceSync 的 kill 请求，内核同步 kill 实例；若为 false，表示向 function-proxy 发送信号量为 killInstance 的 kill 请求，内核异步 kill 实例。

- 抛出：

   - **YRException** - 回收实例失败抛出`YRException`异常，具体错误详见错误信息描述。

#### public Map<String, String> exportHandler() throws YRException

获取实例句柄信息。

CppInstanceHandler 类的成员方法，用户通过该方法可获取句柄信息，可以将其序列化后存入数据库等持久化工具。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
Map<String, String> out = cppInstanceHandler.exportHandler();
// 序列化后存储到数据库或其他持久化工具中。
```

- 返回：

    通过 kv 键值对存放的 CppInstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 导出信息失败抛出 YRException 异常，具体错误详见错误信息描述。

#### public &lt;R&gt; CppInstanceFunctionHandler&lt;R&gt; function(CppInstanceMethod&lt;R&gt; cppInstanceMethod)

CppInstanceHandler 类的成员方法，用于返回云上 C++ 类实例的成员函数的句柄。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
CppInstanceFunctionHandler cppInstFuncHandler = cppInstanceHandler.function(
    CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInstFuncHandler.invoke(5);
int res = (int) YR.get(ref, 100);
```

- 参数：

   - **&lt;R&gt;** - 类型参数。
   - **cppInstanceMethod** - CppInstanceMethod 类实例。

- 返回：

    CppInstanceFunctionHandler 实例。
