# JavaInstanceCreator

包名：`org.yuanrong.call`。

## public class JavaInstanceCreator

创建 Java 有状态函数类实例的操作类。

:::{Note}

JavaInstanceCreator 类是 Java 类实例的创建器，是接口 [YR.instance(JavaInstanceClass javaInstanceClass)](Instance.md)的返回值类型。用户可以使用 JavaInstanceCreator 的 invoke 方法创建 Java 类实例，并返回句柄类 JavaInstanceHandler。

:::

### 接口说明

#### public JavaInstanceCreator(JavaInstanceClass javaInstanceClass)

JavaInstanceCreator 的构造函数。

- 参数：

   - **javaInstanceClass** - JavaInstanceClass 类实例。

#### public JavaInstanceCreator(JavaInstanceClass javaInstanceClass, String name, String nameSpace)

JavaInstanceCreator 的构造函数。

- 参数：

   - **javaInstanceClass** - JavaInstanceClass 类实例。
   - **name** - 具名实例的实例名。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - 具名实例的命名空间。当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name。该字段目前仅用做拼接，namespace 隔离等相关功能待后续补全。

#### public JavaInstanceHandler invoke(Object... args) throws YRException

类的成员方法，用于创建 Java 类实例。

- 参数：

   - **args** - invoke 调用类实例创建所需的入参。

- 返回：

    JavaInstanceHandler 类句柄。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public JavaInstanceCreator setUrn(String urn)

Java 调用 Java 有状态函数时，为函数设置 functionUrn。

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp"))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    JavaInstanceCreator，内置 invoke 方法，可对该 Java 函数类实例进行创建。

#### public JavaInstanceCreator options(InvokeOptions opt)

JavaInstanceCreator 类的成员方法，用于动态修改 Java 函数实例创建的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
JavaInstanceCreator javaInstanceCreator = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
JavaInstanceHandler javaInstanceHandler = javaInstanceCreator.invoke();
ObjectRef ref = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    JavaInstanceCreator 类句柄。

#### public FunctionMeta getFunctionMeta()

JavaInstanceCreator 类的成员方法，用于返回函数元信息。

- 返回：

    FunctionMeta 类实例：函数元信息。
