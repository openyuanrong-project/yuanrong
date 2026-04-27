# JavaInstanceMethod

包名：`org.yuanrong.function`。

## public class JavaInstanceMethod&lt;R\>

用于调用成员函数的辅助类。

### 接口说明

#### public static JavaInstanceMethod\<Object> of(String methodName)

Java 函数调用 Java 类实例成员函数。

- 参数：

   - **methodName** - Java 实例方法名称。

- 返回：

    JavaInstanceMethod 实例。

#### public static &lt;R&gt; JavaInstanceMethod&lt;R&gt; of(String methodName, Class&lt;R&gt; returnType)

Java 函数调用 Java 类实例成员函数。

```java

ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
```

- 参数：

   - **methodName** - Java 实例方法名称。
   - **returnType** - Java 类实例方法的返回值类。
   - **&lt;R&gt;** - Java 类实例方法的返回值类型。

- 返回：

    JavaInstanceMethod 实例。
