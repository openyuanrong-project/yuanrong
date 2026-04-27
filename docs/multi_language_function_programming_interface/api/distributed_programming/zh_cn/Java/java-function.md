# JavaFunction

包名：`org.yuanrong.function`。

## public class JavaFunction&lt;R\>

Java 函数调用 Java 函数。

### 接口说明

#### public static JavaFunction\<Object\> of(String className, String functionName)

Java 函数调用 Java 函数。

- 参数：

   - **className** - Java 类名称。
   - **functionName** - Java 函数名称。

- 返回：

    JavaFunction 实例。

#### public static &lt;R&gt; JavaFunction&lt;R&gt; of(String className, String functionName, Class&lt;R&gt; returnType)

Java 函数调用 Java 函数。

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class)).invoke();
```

- 参数：

   - **className** - Java 类名称。
   - **functionName** - Java 函数名称。
   - **returnType** - Java 方法的返回值类型。
   - **&lt;R&gt;** - 返回值类型。

- 返回：

    JavaFunction 实例。
