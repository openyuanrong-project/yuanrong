# JavaInstanceFunctionHandler

包名：`org.yuanrong.call`。

## public class JavaInstanceFunctionHandler&lt;R>

调用 Java 实例成员函数的操作类。

:::{Note}

类 JavaInstanceFunctionHandler 是创建 Java 类实例后，Java 类实例成员函数的句柄。是接口 [JavaInstanceHandler.function](JavaInstanceHandler.md)的返回值类型。用户可以使用 JavaInstanceFunctionHandler 的 invoke 方法调用 Java 类实例的成员函数。

:::

### 接口说明

#### public ObjectRef invoke(Object... args) throws YRException

JavaInstanceFunctionHandler 类的成员方法，用于调用 Java 类实例的成员函数。

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public JavaInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

JavaInstanceFunctionHandler 类的成员方法，用于动态修改被调用 Java 函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).options(invokeOptions);
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

   JavaInstanceFunctionHandler 类句柄。

#### JavaInstanceFunctionHandler(String instanceId, String functionId, String className, JavaInstanceMethod&lt;R&gt; javaInstanceMethod)

JavaInstanceFunctionHandler 的构造函数。

- 参数：

   - **&lt;R&gt;** - 对象类型。
   - **instanceId** - Java 函数实例 ID。
   - **functionId** - Java 函数部署返回的 ID。
   - **className** - Java 函数所属类名。
   - **javaInstanceMethod** - JavaInstanceMethod 类实例。
