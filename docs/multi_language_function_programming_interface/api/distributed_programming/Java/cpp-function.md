# CppFunction

package: `org.yuanrong.function`.

## public class CppFunction&lt;R\>

Java function calls C++ function.

### Interface description

#### public static CppFunction\<Object\> of(String functionName)

Java function calls C++ function.

- Parameters:

   - **functionName** - C++ method name.

- Returns:

    CppFunction Instance.

#### public static &lt;R&gt; CppFunction&lt;R&gt; of(String functionName, Class&lt;R&gt; returnType)

Java function calls C++ function.

```java

ObjectRef res = YR.function(CppFunction.of("PlusOne", int.class)).invoke(1);
```

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **functionName** - C++ method name.
   - **returnType** - Return type of a C++ method.

- Returns:

    CppFunction Instance.
