# invoke

## 介绍

包名: `org.yuanrong.call`。

为以下实例的内置方法时，将创建所指定的实例：

- InstanceCreator
- CppInstanceCreator

为以下实例的内置方法时，将调用指定方法：

- FunctionHandler/InstanceFunctionHandler
- VoidFunctionHandler/VoidInstanceFunctionHandler
- CppFunctionHandler/CppInstanceFunctionHandler
- JavaFunctionHandler/JavaInstanceFunctionHandler

### 约束

```text
1、由于嵌套引用计数限制，用户在函数间传递 ObjectRef 时，须使用 List.class 类型的容器来承载 ObjectRef （当前不能使用 ArrayList）。ObjectRef 也不能跟其他非 ObjectRef 类型共用一个容器，例如不允许将 List 初始化为 List<Object> 后同时存放 ObjectRef 类型和 Integer 类型等。

2、基于约束 1，cpp 和 java 互调场景请勿使用 list 容器承载非 ObjectRef 对象作为用户函数入参出参，如必须使用 list ，可以使用自定义数据类型对 list 做封装以规避此约束。

3、请勿使用 array 容器（例如 int[] ）作为入参。

4、请勿使用 com.google.gson.JsonObject 类实例作为入参、出参以及返回值。
```

### 接口说明

#### public ObjectRef FunctionHandler.invoke(Object... args) throws YRException

java 函数调用接口。

- 参数：

   - **args** - 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef, 方法返回值在数据系统的 “id” ，使用[YR.get()](get.md)可获取方法的实际返回值。

#### public void VoidFunctionHandler.invoke(Object... args) throws YRException

[VoidFunctionHandler](VoidFunctionHandler.md) 类的成员方法，用于调用 void 函数。

- 参数：

   - **args** - 调用指定方法所需的入参。

- 抛出：
   - **YRException** - 统一抛出的异常类型。

#### public InstanceHandler InstanceCreator.invoke(Object... args) throws YRException

[InstanceCreator](InstanceCreator.md) 类的成员方法，用于创建 java 类实例。

- 参数：

   - **args** - 调用类实例创建所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    [InstanceHandler](InstanceHandler.md) 类句柄。

#### public ObjectRef InstanceFunctionHandler.invoke(Object... args) throws YRException

[InstanceFunctionHandler](InstanceFunctionHandler.md) 类的成员方法，用于调用 Java 类实例的成员函数。

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
InstanceHandler instanceHandler = YR.instance(Counter::new).invoke(1);
InstanceFunctionHandler insFuncHandler = instanceHandler.function(Counter::Add);
ObjectRef ref = insFuncHandler.options(invokeOptions).invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **args** - 调用类实例创建所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

#### public void VoidInstanceFunctionHandler.invoke(Object... args) throws YRException

[VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 类的成员方法，用于调用 void 类实例的成员函数。

- 参数：

   - **args** - 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### public ObjectRef JavaFunctionHandler.invoke(Object... args) throws YRException

java 函数调用接口。

- 参数：

   - **args** - 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

#### public JavaInstanceHandler JavaInstanceCreator.invoke(Object... args) throws YRException

[JavaInstanceCreator](JavaInstanceCreator.md) 类的成员方法，用于创建 java 类实例。

- 参数：

   - **args** - 调用类实例创建所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    [JavaInstanceHandler](JavaInstanceHandler.md) 类句柄。

#### public ObjectRef JavaInstanceFunctionHandler.invoke(Object... args) throws YRException

[JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) 类的成员方法，用于创建 java 类实例。

- 参数：

   - **args** - 调用类实例创建所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

#### public ObjectRef CppFunctionHandler.invoke(Object... args) throws YRException

C++ 函数调用接口。

```java

CppFunctionHandler cppFuncHandler = YR.function(CppFunction.of("Add", int.class));
ObjectRef ref = cppFuncHandler.invoke(1);
int result = YR.get(ref, 15);
```

- Parameters:

   - **args** - 调用类实例创建所需的入参。

- Throws:

   - **YRException** - 统一抛出的异常类型。

- Returns:

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。

#### public CppInstanceHandler CppInstanceCreator.invoke(Object... args) throws YRException

[CppInstanceCreator](CppInstanceCreator.md) 类的成员方法，用于创建 cpp 类实例。

- 参数：

   - **args** - 调用指定方法所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md)类句柄。

#### public ObjectRef CppInstanceFunctionHandler.invoke(Object... args) throws YRException

[CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) 类的成员方法，用于创建 cpp 类实例。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
CppInstanceFunctionHandler cppInsFuncHandler = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInsFuncHandler.invoke(5);
int res = (int)YR.get(ref, 100);
```

- 参数：

   - **args** - 调用类实例创建所需的入参。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

- 返回：

    ObjectRef：方法返回值在数据系统的 “id” ，使用 [YR.get()](get.md) 可获取方法的实际返回值。
