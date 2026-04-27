# JavaInstanceClass

package: `org.yuanrong.function`.

## public class JavaInstanceClass

Helper class to new java instance class.

### Interface description

#### public static JavaInstanceClass of(String className)

Java function creates Java class instance.

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).invoke();
```

- Parameters:

   - **className** - Class name of Java instance.

- Returns:

    JavaInstanceClass Instance.
