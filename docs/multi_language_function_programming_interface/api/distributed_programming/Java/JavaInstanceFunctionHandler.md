# JavaInstanceFunctionHandler

package: `org.yuanrong.call`.

## public class JavaInstanceFunctionHandler&lt;R>

Class that invokes Java instance member functions.

:::{Note}

The JavaInstanceFunctionHandler class is a handle to the member functions of a Java class instance after the Java class instance is created; it is the return type of the interface [JavaInstanceHandler.function](JavaInstanceHandler.md).

Users can use the invoke method of JavaInstanceFunctionHandler to call member functions of Java class instances.

:::

### Interface description

#### public ObjectRef invoke(Object... args) throws YRException

The member method of the JavaInstanceFunctionHandler class is used to call the member function of a Java class instance.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Returns:

   ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public JavaInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

The member method of the JavaInstanceFunctionHandler class is used to dynamically modify the parameters of the called Java function.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).options(invokeOptions);
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

   JavaInstanceFunctionHandler Class handle.

#### JavaInstanceFunctionHandler(String instanceId, String functionId, String className, JavaInstanceMethod&lt;R&gt; javaInstanceMethod)

The constructor of JavaInstanceFunctionHandler.

- Parameters:

   - **&lt;R&gt;** - the type of the object.
   - **instanceId** - Java function instance ID.
   - **functionId** - Java function deployment returns an ID.
   - **className** - Java function class name.
   - **javaInstanceMethod** - JavaInstanceMethod class instance.
