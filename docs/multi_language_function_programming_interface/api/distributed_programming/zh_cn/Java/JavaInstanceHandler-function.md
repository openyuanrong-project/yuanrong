# JavaInstanceHandler function

包名：`org.yuanrong.function`。

## public class JavaInstanceHandler

创建 Java 有状态函数实例的操作类。

:::{Note}

类 [JavaInstanceHandler](JavaInstanceHandler.md) 是创建 Java 类实例后返回的句柄。 是接口 [JavaInstanceCreator.invoke](JavaInstanceCreator.md) 创建 Java 类实例后的返回值类型。

用户可以使用 [JavaInstanceHandler](JavaInstanceHandler.md) 的 function 方法创建 Java 类实例成员方法句柄，并返回句柄类 [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md).

:::

### 接口说明

#### public <R> JavaInstanceFunctionHandler<R> function(JavaInstanceMethod<R> javaInstanceMethod)

JavaInstanceHandler 类的成员方法，用于返回云上 Java 类实例的成员函数句柄。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class))
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **<R>** - object 的类型。
   - **javaInstanceMethod** - JavaInstanceMethod 类实例。

- 返回：

    JavaInstanceFunctionHandler 实例。
