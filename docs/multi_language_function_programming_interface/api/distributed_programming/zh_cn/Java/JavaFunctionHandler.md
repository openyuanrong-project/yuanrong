# JavaFunctionHandler

包名：`org.yuanrong.call`。

## public class JavaFunctionHandler&lt;R>

创建 Java 无状态函数实例的操作类。

:::{Note}

类 JavaFunctionHandler 是创建 Java 函数实例的句柄。是接口 [function (JavaFunction&lt;R&gt; javaFunction)](function.md) 的返回值类型。用户可以使用 JavaFunctionHandler 创建并调用 Java 函数实例。

:::

### 接口说明

#### public JavaFunctionHandler(JavaFunction&lt;R&gt; func)

JavaFunctionHandler 构造函数。

- 参数：

   - **func** - JavaFunction 类实例。

#### public ObjectRef invoke(Object... args) throws YRException

Java 函数调用接口。

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef, 方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

#### public JavaFunctionHandler&lt;R&gt; setUrn(String urn)

java 调用 java 无状态函数时，为函数设置 functionUrn。

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
String res = (String)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    JavaFunctionHandler&lt;R&gt;, 内置 invoke 方法，可对该 Java 函数实例进行创建并调用。
