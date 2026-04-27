# VoidInstanceFunctionHandler

package: `org.yuanrong.call`.

## public class VoidInstanceFunctionHandler extends Handler

Operation class that calls a Java class instance member function without a return value.

:::{Note}

The VoidInstanceFunctionHandler class is a handle to the member function of a class instance after creating a Java class instance without a return value; it is the return type of the interface [InstanceHandler.function](InstanceHandler.md).

Users can use the invoke method of VoidInstanceFunctionHandler to call member functions of Java class instances without return values.

:::

### Interface description

#### public VoidInstanceFunctionHandler(YRFuncVoid func, String instanceId, ApiType apiType)

The constructor of VoidInstanceFunctionHandler.

- Parameters:

   - **func** - Java function name, supports 0 ~ 5 parameters, no return value user function.
   - **instanceId** - Java function instance ID.
   - **apiType** - The enumeration class has two values: Function and Posix. It is used internally by openYuanrong to distinguish function types. The default is Function.

#### public void invoke(Object... args) throws YRException

Member method of the VoidInstanceFunctionHandler class, used to call member functions of void class instances.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public VoidInstanceFunctionHandler options(InvokeOptions options)

Member method of the VoidInstanceFunctionHandler class, used to dynamically modify the parameters of the called void function.

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

   - **options** - Function call options, used to specify functions such as calling resources.

- Returns:

    VoidInstanceFunctionHandler Class handle.
