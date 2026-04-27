# JavaInstanceClass

包名：`org.yuanrong.function`。

## public class JavaInstanceClass

用于创建 Java 实例类的辅助类。

### 接口说明

#### public static JavaInstanceClass of(String className)

Java 函数创建 Java 类实例。

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).invoke();
```

- 参数：

   - **className** - Java 实例的类名。

- 返回：

    JavaInstanceClass 实例。
