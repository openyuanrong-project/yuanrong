# JavaInstanceHandler function

package: `org.yuanrong.function`.

## public class JavaInstanceHandler

Create an operation class for creating a Java stateful function instance.

:::{Note}

The [JavaInstanceHandler](JavaInstanceHandler.md) class is the handle returned after creating a Java class instance; it is the return type of the [JavaInstanceCreator.invoke](JavaInstanceCreator.md) interface after creating a Java class instance.

Users can use the function method of [JavaInstanceHandler](JavaInstanceHandler.md) to create a Java class instance member method handle and return the handle class [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md).

:::

### Interface description

#### public <R> JavaInstanceFunctionHandler<R> function(JavaInstanceMethod<R> javaInstanceMethod)

The member method of the JavaInstanceHandler class is used to return the member function handle of the Java class instance on the cloud.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class))
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- Parameters:

   - **<R>** - the type of the object.
   - **javaInstanceMethod** - JavaInstanceMethod Class instance.

- Returns:

    JavaInstanceFunctionHandler Instance.
