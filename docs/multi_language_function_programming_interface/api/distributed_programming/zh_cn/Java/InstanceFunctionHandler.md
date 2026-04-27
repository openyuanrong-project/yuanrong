# InstanceFunctionHandler

包名：`org.yuanrong.call`。

## public class InstanceFunctionHandler&lt;R&gt; extends Handler

调用 java 有状态函数实例成员函数的操作类。

:::{Note}

类 InstanceFunctionHandler 是创建 Java 类实例后，类实例成员函数的句柄，是接口 [InstanceHandler.function](InstanceHandler.md) 的返回值类型。

用户可以使用 InstanceFunctionHandler 的 invoke 方法调用 Java 类实例的成员函数。

:::

### 接口说明

#### public InstanceFunctionHandler(YRFuncR&lt;R&gt; func, String instanceId, ApiType apiType)

InstanceFunctionHandler 的构造函数。

- 参数：

   - **func** – YRFuncR 类实例。
   - **instanceId** – Java 函数实例 ID。
   - **apiType** – 枚举类，具有 Function，Posix 2 个值，openYuanrong 内部用于区分函数类型。默认为 Function。

#### public ObjectRef invoke(Object... args) throws YRException

InstanceFunctionHandler 类的成员方法，用于调用 Java 类实例的成员函数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
InstanceHandler instanceHandler = YR.instance(Counter::new).invoke(1);
InstanceFunctionHandler insFuncHandler = instanceHandler.function(Counter::Add);
ObjectRef ref = insFuncHandler.options(invokeOptions).invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **args** – invoke 调用指定方法所需的入参。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public InstanceFunctionHandler&lt;R&gt; options(InvokeOptions options)

InstanceFunctionHandler 类的成员方法，用于动态修改被调用 Java 函数的参数。

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

   - **options** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    InstanceFunctionHandler 类句柄。
