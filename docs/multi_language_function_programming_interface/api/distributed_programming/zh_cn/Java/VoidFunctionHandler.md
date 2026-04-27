# VoidFunctionHandler

包名：`org.yuanrong.call`。

## public class VoidFunctionHandler extends Handler

调用 void 无返回值函数的操作类。

:::{Note}

类 VoidFunctionHandler 是 Java 无返回值函数实例的句柄。是接口 [YR.function(YRFuncVoid func)](function.md)的返回值类型。用户可以使用 VoidFunctionHandler 创建并调用 Java 无返回值函数实例。

:::

### 接口说明

#### public VoidFunctionHandler(YRFuncVoid func)

VoidFunctionHandler 的构造函数。

- 参数：

   - **func** - Java 函数名, 支持 0 ~ 5 个入参，无返回值的用户函数。

#### public void invoke(Object... args) throws YRException

VoidFunctionHandler 类的成员方法，用于调用 void 函数。

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public VoidFunctionHandler options(InvokeOptions opt)

VoidFunctionHandler 类的成员方法，用于动态修改被调用 void 函数的参数。

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

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    VoidFunctionHandler 类句柄。
