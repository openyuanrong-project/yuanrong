# CppInstanceHandler function

包名：`org.yuanrong.call`。

## public class CppInstanceHandler

创建 cpp 有状态函数实例的操作类。

:::{Note}

类 [CppInstanceHandler](CppInstanceHandler.md) 是 Java 函数创建 cpp 类实例后返回的句柄。是接口 [CppInstanceCreator.invoke](CppInstanceCreator.md) 创建 cpp 类实例后的返回值类型。 用户可以使用 [CppInstanceHandler](CppInstanceHandler.md) 的 function 方法创建 cpp 类实例成员方法句柄，并返回句柄类 [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md)。

:::

### 接口说明

#### public &lt;R&gt; CppInstanceFunctionHandler&lt;R&gt; function(CppInstanceMethod&lt;R&gt; cppInstanceMethod)

[CppInstanceHandler](CppInstanceHandler.md) 类的成员方法，用于返回云上 C++ 类实例的成员函数的句柄。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
CppInstanceFunctionHandler cppInstFuncHandler = cppInstanceHandler.function(
    CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInstFuncHandler.invoke(5);
int res = (int) YR.get(ref, 100);
```

- 参数：

   - **&lt;R&gt;** - 参数类型。
   - **cppInstanceMethod** - CppInstanceMethod 类实例。

- 返回：

    [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) 实例。
