# JavaFunction

package: `org.yuanrong.function`.

## public class JavaFunction&lt;R\>

Java function calls Java function.

### Interface description

#### public static JavaFunction\<Object\> of(String className, String functionName)

Java function calls Java function.

- Parameters:

   - **className** - Java class name.
   - **functionName** - Java function name.

- Returns:

    JavaFunction Instance.

#### public static &lt;R&gt; JavaFunction&lt;R&gt; of(String className, String functionName, Class&lt;R&gt; returnType)

Java function calls Java function.

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class)).invoke();
```

- Parameters:

   - **className** - Java class name.
   - **functionName** - Java function name.
   - **returnType** - The return type of a Java method.
   - **&lt;R&gt;** - Return value type.

- Returns:

    JavaFunction Instance.
