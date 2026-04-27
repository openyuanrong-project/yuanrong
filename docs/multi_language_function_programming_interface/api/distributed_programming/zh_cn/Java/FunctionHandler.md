# FunctionHandler

包名：`org.yuanrong.call`。

## public class FunctionHandler<R> extends Handler

创建 java 无状态函数实例的操作类。

:::{Note}

类 FunctionHandler 是云上 Java 函数的句柄，是接口 [YR.function(YRFuncR&lt;R&gt; func)](function.md) 的返回值类型。

用户可以使用 FunctionHandler 创建并调用 Java 函数实例。

:::

### 接口说明

#### public FunctionHandler(YRFuncR&lt;R&gt; func)

FunctionHandler 构造函数。

- 参数：

   - **func** - Java 函数名, openYuanrong 当前支持 0 ~ 5 个入参，一个返回值的用户函数。

#### public ObjectRef invoke(Object... args) throws YRException

Java 函数调用接口。

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 返回：

    ObjectRef, 方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public FunctionHandler&lt;R&gt; options(InvokeOptions opt)

用于动态修改被调用函数的参数。

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

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    FunctionHandler 类句柄。
