# InstanceCreator

包名：`org.yuanrong.call`。

## public class InstanceCreator<A> extends Handler

创建 Java 有状态函数实例的操作类。

:::{Note}

类 instanceCreator 是 Java 类实例的创建者，是接口 [YR.instance(YRFuncR&lt;A&gt; func)](Instance.md) 的返回值类型。

用户可以使用 instanceCreator 的 invoke 方法创建 Java 类实例，并返回句柄类[InstanceHandler](InstanceHandler.md)。

:::

### 接口说明

#### public InstanceCreator(YRFuncR&lt;A&gt; func)

InstanceCreator 的构造函数。

- 参数：

   - **func** - Java 函数名。YRFuncR 类实例。

#### public InstanceCreator(YRFuncR&lt;A&gt; func, ApiType apiType)

InstanceCreator 的构造函数。

- 参数：

   - **func** - Java 函数名。YRFuncR 类实例。
   - **apiType** - 枚举类，具有 Function，Posix 2 个值，openYuanrong 内部用于区分函数类型。默认为 Function。

#### public InstanceCreator(YRFuncR&lt;A&gt; func, String name, String nameSpace)

InstanceCreator 的构造函数。

- 参数：

   - **func** - Java 函数名。YRFuncR 类实例。
   - **name** - 具名实例的实例名。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - 具名实例的命名空间。当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name。该字段目前仅用做拼接，namespace 隔离等相关功能待后续补全。

#### public InstanceCreator(YRFuncR&lt;A&gt; func, String name, String nameSpace, ApiType apiType)

InstanceCreator 的构造函数。

- 参数：

   - **func** - Java 函数名。YRFuncR 类实例。
   - **name** - 具名实例的实例名。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - 具名实例的命名空间。当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name。该字段目前仅用做拼接，namespace 隔离等相关功能待后续补全。
   - **apiType** - 枚举类，具有 Function，Posix 2 个值，openYuanrong 内部用于区分函数类型。默认为 Function。

#### public InstanceHandler invoke(Object... args) throws YRException

InstanceCreator 类的成员方法，用于创建 java 类实例。

- 参数：

   - **args** - invoke 调用类实例创建所需的入参。

- 返回：

    [InstanceHandler](InstanceHandler.md) 类句柄。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public InstanceCreator&lt;A&gt; options(InvokeOptions opt)

InstanceCreator 类的成员方法，用于动态修改 java 函数实例创建的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
InstanceCreator instanceCreator = YR.instance(Counter::new).options(invokeOptions);
InstanceHandler instanceHandler = instanceCreator.invoke(1);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    InstanceCreator 类句柄。
