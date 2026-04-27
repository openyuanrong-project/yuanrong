# CppInstanceClass

package: `org.yuanrong.function`.

## public class CppInstanceClass

Helper class to new cpp instance class.

### Interface description

#### public static CppInstanceClass of(String className, String initFunctionName)

Java function creates C++ class instance.

```java

CppInstanceClass cppInstance = CppInstanceClass.of("Counter", "FactoryCreate");
```

- Parameters:

   - **className** - Class name of C++ instance.
   - **initFunctionName** – C++ class initialization function.

- Returns:

    CppInstanceClass instance.
