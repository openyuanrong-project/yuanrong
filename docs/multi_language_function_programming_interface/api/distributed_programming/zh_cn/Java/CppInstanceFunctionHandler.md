# CppInstanceFunctionHandler

包名：`org.yuanrong.call`。

## public class CppInstanceFunctionHandler&lt;R>

调用 cpp 有状态函数实例成员函数的操作类。

:::{Note}

类 CppInstanceFunctionHandler 是 Java 函数创建 c++ 类实例后，c++ 类实例成员函数的句柄。是接口 [CppInstanceHandler.function](CppInstanceHandler.md) 的返回值类型。

用户可以使用 CppInstanceFunctionHandler 的 invoke 方法调用 c++ 类实例的成员函数。

:::

- 参数：

   - **&lt;R&gt;** - 成员函数的具体类型。

### 接口说明

#### public ObjectRef invoke(Object... args) throws YRException

CppInstanceFunctionHandler 类的成员方法，用于调用 cpp 类实例的成员函数。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
CppInstanceFunctionHandler cppInsFuncHandler = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInsFuncHandler.invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **args** - invoke 调用指定方法所需的入参。

- 返回：

    ObjectRef，方法返回值在数据系统的 "id" ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public CppInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

CppInstanceFunctionHandler 类的成员方法，用于动态修改被调用函数的参数。

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    CppInstanceFunctionHandler 类句柄。

#### CppInstanceFunctionHandler(String instanceId, String functionId, String className, CppInstanceMethod&lt;R&gt; cppInstanceMethod)

CppInstanceFunctionHandler 的构造函数。

- 参数：

   - **instanceId** - cpp 函数实例 ID。
   - **functionId** - cpp 函数部署返回的 ID。
   - **className** - cpp 函数所属类名。
   - **cppInstanceMethod** - cppInstanceMethod 类实例。
