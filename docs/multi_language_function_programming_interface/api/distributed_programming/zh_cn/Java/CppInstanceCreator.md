# CppInstanceCreator

包名：`org.yuanrong.call`。

## public class CppInstanceCreator

创建 cpp 有状态函数类实例的操作类。

:::{Note}

CppInstanceCreator 类是一个用于创建 C++ 类实例的 Java 函数，是接口 [YR.instance(CppInstanceClass cppInstanceClass)](Instance.md)的返回值类型。

用户可以使用 CppInstanceCreator 的 invoke 方法创建 c++ 类实例，并返回句柄类 CppInstanceHandler。

:::

### 接口说明

#### public CppInstanceCreator(CppInstanceClass cppInstanceClass)

CppInstanceCreator 的构造函数。

- 参数：

   - **cppInstanceClass** - CppInstanceClass 类实例。

#### public CppInstanceCreator(CppInstanceClass cppInstanceClass, String name, String nameSpace)

CppInstanceCreator 的构造函数。

- 参数：

   - **cppInstanceClass** - CppInstanceClass 类实例。
   - **name** - 具名实例的实例名。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - 具名实例的命名空间。当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name。该字段目前仅用做拼接，namespace 隔离等相关功能待后续补全。

#### public CppInstanceHandler invoke(Object... args) throws YRException

CppInstanceHandler 类的成员方法，用于创建 cpp 类实例。

- 参数：

   - **args** - invoke 调用类实例创建所需的入参。

- 返回：

    [CppInstanceHandler](CppInstanceHandler.md) 类句柄。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public CppInstanceCreator setUrn(String urn)

java 调用 cpp 有状态函数时，为函数设置 functionUrn。

```java

CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter","FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    CppInstanceCreator，内置 invoke 方法，可对该 cpp 函数类实例进行创建。

#### public CppInstanceCreator options(InvokeOptions opt)

CppInstanceCreator 类的成员方法，用于动态修改 cpp 函数实例创建的参数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
CppInstanceCreator cppInstanceCreator = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
CppInstanceHandler cppInstanceHandler = cppInstanceCreator.invoke(1);
ObjectRef ref = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **opt** - 函数调用选项，用于指定调用资源等功能。

- 返回：

    CppInstanceCreator 类句柄。
