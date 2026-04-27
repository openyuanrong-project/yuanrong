# function options

## function.options()

package: `org.yuanrong.call`.

### Interface description

#### public FunctionHandler&lt;R&gt; options(InvokeOptions opt)

The member method of the [FunctionHandler](FunctionHandler.md) class is used to dynamically modify the parameters of the called function.

```java

Config conf = new Config("FunctionURN", "ip", "ip", "", false);
YR.init(conf);
InvokeOptions opts = new InvokeOptions();
opts.setCpu(300);
FunctionHandler<String> f_h = YR.function(MyYRApp::myFunction).options(opts);
ObjectRef res = f_h.invoke();
System.out.println("myFunction invoke ref:" + res.getObjId());
YR.Finalize();
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [FunctionHandler](FunctionHandler.md) Class handle.

#### public VoidFunctionHandler options(InvokeOptions opt)

The member method of the [VoidFunctionHandler](VoidFunctionHandler.md) class is used to dynamically modify the parameters of the called void function.

```java

Config conf = new Config("FunctionURN", "ip", "ip", "", false);
YR.init(conf);

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
VoidFunctionHandler<String> functionHandler = YR.function(MyYRApp::myVoidFunction).options(invokeOptions);
functionHandler.invoke();
YR.Finalize();
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [VoidFunctionHandler](VoidFunctionHandler.md) Class handle.

#### public InstanceCreator&lt;A&gt; options(InvokeOptions opt)

The member method of the [InstanceCreator](InstanceCreator.md) class is used to dynamically modify the parameters for creating a Java function instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
InstanceCreator instanceCreator = YR.instance(Counter::new).options(invokeOptions);
InstanceHandler instanceHandler = instanceCreator.invoke(1);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [InstanceCreator](InstanceCreator.md) Class handle.

#### public InstanceFunctionHandler&lt;R&gt; options(InvokeOptions options)

The member method of the [InstanceFunctionHandler](InstanceFunctionHandler.md) class is used to dynamically modify the parameters of the called Java function.

```java

YR.init(conf);
InvokeOptions options = new InvokeOptions();
options.setConcurrency(100);
options.getCustomResources().put("nvidia.com/gpu", 100F);
InstanceCreator<MyYRApp> f_instance = YR.instance(MyYRApp::new);
InstanceHandler f_myHandler = f_instance.invoke();
InstanceFunctionHandler<String> f_h = f_myHandler.function(MyYRApp::myFunction);
ObjectRef res = f_h.options(options).invoke();
System.out.println("myFunction invoke ref:" + res.getObjId());
//expected result: "myFunction invoke ref: obj-***-***"
YR.Finalize();
```

- Parameters:

   - **options** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Class handle.

#### public VoidInstanceFunctionHandler options(InvokeOptions options)

Member method of the [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) class, used to dynamically modify the parameters of the called void function.

```java

Config conf = new Config("FunctionURN", "ip", "ip", "", false);
YR.init(conf);

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
InstanceHandler instanceHandler = YR.instance(MyYRApp::new).invoke();
VoidInstanceFunctionHandler insFuncHandler = instanceHandler.function(MyYRApp::myVoidFunction).options(invokeOptions);
insFuncHandler.invoke();
YR.Finalize();
```

- Parameters:

   - **options** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Class handle.

#### public JavaFunctionHandler&lt;R&gt; options(InvokeOptions opt)

The member method of the [JavaFunctionHandler](JavaFunctionHandler.md) class is used to dynamically modify the parameters of the called function.

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

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    JavaFunctionHandler&lt;R&gt;, Handle processed by Java function.

#### public JavaInstanceCreator options(InvokeOptions opt)

The member method of the [JavaInstanceCreator](JavaInstanceCreator.md) class is used to dynamically modify the parameters for creating a Java function instance.

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

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [JavaInstanceCreator](JavaInstanceCreator.md)Class handle.

#### public JavaInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

The member method of the [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) class is used to dynamically modify the parameters of the called Java function.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).options(invokeOptions);
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) Class handle.

#### public CppFunctionHandler&lt;R&gt; options(InvokeOptions opt)

The member method of the [CppFunctionHandler](CppFunctionHandler.md) class is used to dynamically modify the parameters of the called function.

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    CppFunctionHandler&lt;R&gt;, Handles processed by C++ functions.

#### public CppInstanceCreator options(InvokeOptions opt)

The member method of the [CppInstanceCreator](CppInstanceCreator.md) class is used to dynamically modify the parameters for creating a C++ function instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
CppInstanceCreator cppInstanceCreator = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
CppInstanceHandler cppInstanceHandler = cppInstanceCreator.invoke(1);
ObjectRef ref = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [CppInstanceCreator](CppInstanceCreator.md) Class handle.

#### public CppInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

Member method of the [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) class, used to dynamically modify the parameters of the called function.

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) Class handle.
