# JavaInstanceMethod

package: `org.yuanrong.function`.

## public class JavaInstanceMethod&lt;R\>

Helper class to invoke member function.

### Interface description

#### public static JavaInstanceMethod\<Object\> of(String methodName)

Java function calls Java class instance member function.

- Parameters:

   - **methodName** - Java instance method name.

- Returns:

    JavaInstanceMethod Instance.

#### public static &lt;R&gt; JavaInstanceMethod&lt;R&gt; of(String methodName, Class&lt;R&gt; returnType)

Java function calls Java class instance member function.

```java

ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
```

- Parameters:

   - **methodName** - Java instance method name.
   - **returnType** - Java class instance method return value class.
   - **&lt;R&gt;** - The return type of a Java class instance method.

- Returns:

    JavaInstanceMethod Instance.
