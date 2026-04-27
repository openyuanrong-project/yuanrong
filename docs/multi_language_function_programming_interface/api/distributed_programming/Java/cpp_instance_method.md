# CppInstanceMethod

package: `org.yuanrong.function`.

## public class CppInstanceMethod&lt;R\>

Helper class to invoke member function.

### Interface description

#### public static CppInstanceMethod\<Object\> of(String methodName)

Java function calls C++ class instance member function.

- Parameters:

   - **methodName** - C++ instance method name.

- Returns:

    Object Instance of type CppInstanceMethod.

#### public static &lt;R&gt; CppInstanceMethod&lt;R&gt; of(String methodName, Class&lt;R&gt; returnType)

Java function calls C++ class instance member function.

```java

ObjectRef<int> res = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(1);
```

- Parameters:

   - **methodName** - C++ instance method name.
   - **returnType** - Return type of C++ class instance method.
   - **&lt;R&gt;** - The type of the return value of the instance method.

- Returns:

    CppInstanceMethod Instance.
