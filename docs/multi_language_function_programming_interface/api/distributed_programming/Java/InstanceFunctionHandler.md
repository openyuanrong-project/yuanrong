# InstanceFunctionHandler

package: `org.yuanrong.call`.

## public class InstanceFunctionHandler&lt;R&gt; extends Handler

Operation class that calls a Java stateful function instance member function.

:::{Note}

The class InstanceFunctionHandler is the handle of the member function of the class instance after the Java class instance is created; it is the return value type of the interface [InstanceHandler.function](InstanceHandler.md).

Users can use the invoke method of InstanceFunctionHandler to call member functions of Java class instances.

:::

### Public Functions description

#### public InstanceFunctionHandler(YRFuncR&lt;R&gt; func, String instanceId, ApiType apiType)

The constructor of the InstanceFunctionHandler.

- Parameters:

   - **func** – YRFuncR Class instance.
   - **instanceId** – Java function instance ID.
   - **apiType** – The enumeration class has two values: Function and Posix. It is used internally by openYuanrong to distinguish function types. The default is Function.

#### public ObjectRef invoke(Object... args) throws YRException

The member method of the InstanceFunctionHandler class is used to call the member function of a Java class instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
InstanceHandler instanceHandler = YR.instance(Counter::new).invoke(1);
InstanceFunctionHandler insFuncHandler = instanceHandler.function(Counter::Add);
ObjectRef ref = insFuncHandler.options(invokeOptions).invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **args** – The input parameters required to call the specified method.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public InstanceFunctionHandler&lt;R&gt; options(InvokeOptions options)

The member method of the InstanceFunctionHandler class is used to dynamically modify the parameters of the called Java function.

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

   - **options** - Function call options, used to specify functions such as calling resources.

- Returns:

    InstanceFunctionHandler Class handle.
