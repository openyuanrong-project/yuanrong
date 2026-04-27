# function options

## function.options()

包名：`org.yuanrong.call`。

### 接口说明

#### public FunctionHandler&lt;R&gt; options(InvokeOptions opt)

[FunctionHandler](FunctionHandler.md) 类的成员方法，用于动态修改被调用函数的参数。

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

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [FunctionHandler](FunctionHandler.md) 类句柄。

#### public VoidFunctionHandler options(InvokeOptions opt)

[VoidFunctionHandler](VoidFunctionHandler.md) 类的成员方法，用于动态修改被调用 void 函数的参数。

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

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [VoidFunctionHandler](VoidFunctionHandler.md) 类句柄。

#### public InstanceCreator&lt;A&gt; options(InvokeOptions opt)

[InstanceCreator](InstanceCreator.md) 类的成员方法，用于动态修改被调用 java 函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
InstanceCreator instanceCreator = YR.instance(Counter::new).options(invokeOptions);
InstanceHandler instanceHandler = instanceCreator.invoke(1);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类句柄。

#### public InstanceFunctionHandler&lt;R&gt; options(InvokeOptions options)

[InstanceFunctionHandler](InstanceFunctionHandler.md) 类的成员方法，用于动态修改被调用 java 函数的参数。

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
//预期结果："myFunction invoke ref: obj-***-***"
YR.Finalize();
```

- 参数：

   - **options** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 类句柄。

#### public VoidInstanceFunctionHandler options(InvokeOptions options)

[VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 类的成员方法，用于动态修改被调用 void 函数的参数。

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

- 参数：

   - **options** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 类句柄。

#### public JavaFunctionHandler&lt;R&gt; options(InvokeOptions opt)

[JavaFunctionHandler](JavaFunctionHandler.md) 类的成员方法，用于动态修改被调用函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
JavaFunctionHandler javaFuncHandler = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
                .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest");
ObjectRef ref = javaFuncHandler.options(invokeOptions).invoke();
String result = (String)YR.get(ref, 15);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [JavaFunctionHandler](JavaFunctionHandler.md) 类句柄。

#### public JavaInstanceCreator options(InvokeOptions opt)

[JavaInstanceCreator](JavaInstanceCreator.md) 类的成员方法，用于动态修改被调用 java 函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
JavaInstanceCreator javaInstanceCreator = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
JavaInstanceHandler javaInstanceHandler = javaInstanceCreator.invoke();
ObjectRef ref = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [JavaInstanceCreator](JavaInstanceCreator.md) 类句柄。

#### public JavaInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

[JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) 类的成员方法，用于动态修改被调用 java 函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
JavaInstanceFunctionHandler javaInstFuncHandler = javaInstanceHandler.function(JavaInstanceMethod.of("smallCall", String.class)).options(invokeOptions);
ObjectRef ref = javaInstFuncHandler.invoke();
String res = (String)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) 类句柄。

#### public CppFunctionHandler&lt;R&gt; options(InvokeOptions opt)

[CppFunctionHandler](CppFunctionHandler.md) 类的成员方法，用于动态修改被调用函数的参数。

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [CppFunctionHandler](CppFunctionHandler.md) 类句柄。

#### public CppInstanceCreator options(InvokeOptions opt)

[CppInstanceCreator](CppInstanceCreator.md) 类的成员方法，用于动态修改被调用 C++ 函数的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
CppInstanceCreator cppInstanceCreator = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
CppInstanceHandler cppInstanceHandler = cppInstanceCreator.invoke(1);
ObjectRef ref = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md) 类句柄。

#### public CppInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

[CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) 类的成员方法，用于动态修改被调用 C++ 函数的参数。

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) 类句柄。
