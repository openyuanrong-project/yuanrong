# JavaInstanceCreator

package: `org.yuanrong.call`.

## public class JavaInstanceCreator

Creates an operation class for creating a Java stateful function instance.

:::{Note}

The JavaInstanceCreator class is a creator of Java class instances; it is the return type of the interface [YR.instance(JavaInstanceClass javaInstanceClass)](Instance.md).

Users can use the invoke method of JavaInstanceCreator to create Java class instances and return the handle class JavaInstanceHandler.

:::

### Interface description

#### public JavaInstanceCreator(JavaInstanceClass javaInstanceClass)

The constructor of JavaInstanceCreator.

- Parameters:

   - **javaInstanceClass** - JavaInstanceClass class instance.

#### public JavaInstanceCreator(JavaInstanceClass javaInstanceClass, String name, String nameSpace)

The constructor of JavaInstanceCreator.

- Parameters:

   - **javaInstanceClass** - JavaInstanceClass class instance.
   - **name** - The instance name of a named instance. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name will be concatenated into nameSpace-name.
   This field is currently only used for concatenation, and namespace isolation and other related functions will be completed later.

#### public JavaInstanceHandler invoke(Object... args) throws YRException

The member method of the JavaInstanceCreator class is used to create a Java class instance.

- Parameters:

   - **args** - The input parameters required to create an instance of the class.

- Returns:

    JavaInstanceHandler Class handle.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public JavaInstanceCreator setUrn(String urn)

When Java calls a Java stateful function, set the functionUrn for the function.

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp"))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    JavaInstanceCreator, with built-in invoke method, can create instances of this Java function class.

#### public JavaInstanceCreator options(InvokeOptions opt)

The member method of the JavaInstanceCreator class is used to dynamically modify the parameters for creating a Java function instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
JavaInstanceCreator javaInstanceCreator = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
JavaInstanceHandler javaInstanceHandler = javaInstanceCreator.invoke();
ObjectRef ref = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref, 100);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    JavaInstanceCreator Class handle.

#### public FunctionMeta getFunctionMeta()

The member method of the JavaInstanceCreator class is used to return function metadata.

- Returns:

    FunctionMeta class instance: function metadata.
