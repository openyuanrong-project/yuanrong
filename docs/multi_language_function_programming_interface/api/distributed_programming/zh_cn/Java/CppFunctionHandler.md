# CppFunctionHandler

包名：`org.yuanrong.call`。

## public class CppFunctionHandler&lt;R>

创建 cpp 无状态函数实例的操作类。

:::{Note}

类 CppFunctionHandler 是 Java 函数创建的云上 c++ 函数的句柄。是接口 [YR.function(CppFunction&lt;R&gt; func)](function.md) 的返回值类型。用户可以使用 CppFunctionHandler 创建并调用 c++ 函数实例。

:::

### 接口说明

#### public CppFunctionHandler(CppFunction&lt;R&gt; func)

CppFunctionHandler 构造函数

- 参数：

   - **func** -  CppFunction 类实例。

#### public ObjectRef invoke(Object... args) throws YRException

cpp 函数调用接口

```java

CppFunctionHandler cppFuncHandler = YR.function(CppFunction.of("Add", int.class));
ObjectRef ref = cppFuncHandler.invoke(1);
int result = YR.get(ref, 15);
```

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 返回：

    ObjectRef, 方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public CppFunctionHandler&lt;R&gt; setUrn(String urn)

java 调用 cpp 无状态函数时，为函数设置 functionUrn。

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    CppFunctionHandler&lt;R&gt;, 内置 invoke 方法，可对该 cpp 函数实例进行创建并调用。

#### public CppFunctionHandler&lt;R&gt; options(InvokeOptions opt)

用于动态修改被调用函数的参数。

- 参数：

   - **opt** - 详见 [InvokeOptions](InvokeOptions.md)。

- 返回：

    CppFunctionHandler&lt;R&gt;，cpp 函数处理的句柄。
