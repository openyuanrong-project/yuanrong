# VoidFunctionHandler

package: `org.yuanrong.call`.

## public class VoidFunctionHandler extends Handler

Operation class that calls a void function with no return value.

:::{Note}

The VoidFunctionHandler class is a handle for Java functions with no return value; it is the return type of the interface [YR.function(YRFuncVoid func)](function.md).

Users can use VoidFunctionHandler to create and invoke Java functions that do not return values.

:::

### Interface description

#### public VoidFunctionHandler(YRFuncVoid func)

The constructor of VoidFunctionHandler.

- Parameters:

   - **func** - Java function name, supports 0 ~ 5 parameters, no return value user function.

#### public void invoke(Object... args) throws YRException

Member method of the VoidFunctionHandler class, used to call void functions.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public VoidFunctionHandler options(InvokeOptions opt)

The member method of the VoidFunctionHandler class is used to dynamically modify the parameters of the called void function.

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

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    VoidFunctionHandler Class handle.
