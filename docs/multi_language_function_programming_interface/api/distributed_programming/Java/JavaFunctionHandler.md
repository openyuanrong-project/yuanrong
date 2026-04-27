# JavaFunctionHandler

package: `org.yuanrong.call`.

## public class JavaFunctionHandler&lt;R>

Create an operation class for creating a stateless Java function instance.

:::{Note}

The JavaFunctionHandler class is a handle for creating Java function instances; it is the return type of the [function (JavaFunction&lt;R&gt; javaFunction)](function.md)

interface. Users can use JavaFunctionHandler to create and invoke Java function instances.

:::

### Interface description

#### public JavaFunctionHandler(JavaFunction&lt;R&gt; func)

JavaFunctionHandler constructor.

- Parameters:

   - **func** - JavaFunction class instance.

#### public ObjectRef invoke(Object... args) throws YRException

Java function call interface.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public JavaFunctionHandler&lt;R&gt; setUrn(String urn)

When Java calls a stateless function in Java, set the functionUrn for the function.

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
String res = (String)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    JavaFunctionHandler&lt;R&gt;, with built-in invoke method, can create and invoke the java function instance.

#### public JavaFunctionHandler&lt;R&gt; options(InvokeOptions opt)

Used to dynamically modify the parameters of the called function.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
JavaFunctionHandler javaFuncHandler = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
               .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest");
ObjectRef ref = javaFuncHandler.options(invokeOptions).invoke();
String result = (String)YR.get(ref, 15);
```

- Parameters:

   - **opt** - See InvokeOptions for details.

- Returns:

    JavaFunctionHandler&lt;R&gt;, Handles processed by Java function.
