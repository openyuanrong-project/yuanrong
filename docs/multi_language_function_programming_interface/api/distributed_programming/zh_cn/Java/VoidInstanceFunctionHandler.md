# VoidInstanceFunctionHandler

包名：`org.yuanrong.call`。

## public class VoidInstanceFunctionHandler extends Handler

调用 Java 无返回值类实例成员函数的操作类。

:::{Note}

类 VoidInstanceFunctionHandler 是创建 Java 无返回值类实例后，类实例成员函数的句柄。是接口 [InstanceHandler.function](InstanceHandler.md)的返回值类型。用户可以使用 VoidInstanceFunctionHandler 的 invoke 方法调用 Java 无返回值类实例的成员函数。

:::

### 接口说明

#### public VoidInstanceFunctionHandler(YRFuncVoid func, String instanceId, ApiType apiType)

VoidInstanceFunctionHandler 的构造函数。

- 参数：

   - **func** - Java 函数名, 支持 0 ~ 5 个入参，无返回值的用户函数。
   - **instanceId** - Java 函数实例 ID。
   - **apiType** - 枚举类，具有 Function，Posix 2 个值，openYuanrong 内部用于区分函数类型。默认为 Function。

#### public void invoke(Object... args) throws YRException

VoidInstanceFunctionHandler 类的成员方法，用于调用 void 类实例的成员函数。

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public VoidInstanceFunctionHandler options(InvokeOptions options)

VoidInstanceFunctionHandler 类的成员方法，用于动态修改被调用 void 函数的参数。

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

   - **options** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    VoidInstanceFunctionHandler 类句柄。
