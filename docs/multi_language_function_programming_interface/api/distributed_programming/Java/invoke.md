# invoke

## Introduction

package: `org.yuanrong.call`.

When acting as a built-in method for the following instances, the specified instance will be created:

- InstanceCreator
- CppInstanceCreator
- JavaInstanceCreator

When acting as a built-in method for the following instances, The specified method will be called:

- FunctionHandler/InstanceFunctionHandler
- VoidFunctionHandler/VoidInstanceFunctionHandler
- CppFunctionHandler/CppInstanceFunctionHandler
- JavaFunctionHandler/JavaInstanceFunctionHandler

### Constraint

```text
1. Due to the nested reference counting limitation, when passing ObjectRef between functions,users must use a container of type List.class to hold ObjectRef (ArrayList cannot be used at present). ObjectRef cannot share a container with other non-ObjectRef types, for example, it is not allowed to initialize a List as List<Object> and store both ObjectRef and Integer types at the same time.

2. Based on constraint 1, do not use the list container to carry non-ObjectRef objects as user function parameters in the cpp and java interoperability scenario. If you must use list, you can use a custom data type to encapsulate the list to avoid this constraint.

3. Do not use array containers (such as int[]) as input parameters.

4. Do not use com.google.gson.JsonObject class instances as input parameters, output parameters, and return values.
```

### Interface description

#### public ObjectRef FunctionHandler.invoke(Object... args) throws YRException

Java function call interface.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - the YR exception.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public void VoidFunctionHandler.invoke(Object... args) throws YRException

Member method of the [VoidFunctionHandler](VoidFunctionHandler.md) class, used to call void functions.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public InstanceHandler InstanceCreator.invoke(Object... args) throws YRException

The member method of the [InstanceCreator](InstanceCreator.md) class is used to create a Java class instance.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    [InstanceHandler](InstanceHandler.md) Class handle.

#### public ObjectRef InstanceFunctionHandler.invoke(Object... args) throws YRException

The member method of the [InstanceFunctionHandler](InstanceFunctionHandler.md) class is used to call the member function of a Java class instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.addCustomExtensions("app_name", "myApp");
InstanceHandler instanceHandler = YR.instance(Counter::new).invoke(1);
InstanceFunctionHandler insFuncHandler = instanceHandler.function(Counter::Add);
ObjectRef ref = insFuncHandler.options(invokeOptions).invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public void VoidInstanceFunctionHandler.invoke(Object... args) throws YRException

Member method of the [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) class, used to call member functions of void class instances.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public ObjectRef JavaFunctionHandler.invoke(Object... args) throws YRException

Java function call interface.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public JavaInstanceHandler JavaInstanceCreator.invoke(Object... args) throws YRException

The member method of the [JavaInstanceCreator](JavaInstanceCreator.md) class is used to create a Java class instance.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    [JavaInstanceHandler](JavaInstanceHandler.md) Class handle.

#### public ObjectRef JavaInstanceFunctionHandler.invoke(Object... args) throws YRException

The member method of the [JavaInstanceFunctionHandler](JavaInstanceFunctionHandler.md) class is used to call the member function of a Java class instance.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public ObjectRef CppFunctionHandler.invoke(Object... args) throws YRException

Cpp function call interface.

```java

CppFunctionHandler cppFuncHandler = YR.function(CppFunction.of("Add", int.class));
ObjectRef ref = cppFuncHandler.invoke(1);
int result = YR.get(ref, 15);
```

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

#### public CppInstanceHandler CppInstanceCreator.invoke(Object... args) throws YRException

The member method of the [CppInstanceCreator](CppInstanceCreator.md) class is used to create a cpp class instance.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    [CppInstanceCreator](CppInstanceCreator.md) Class handle.

#### public ObjectRef CppInstanceFunctionHandler.invoke(Object... args) throws YRException

The member method of the [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) class is used to call the member function of a cpp class instance.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
CppInstanceFunctionHandler cppInsFuncHandler = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInsFuncHandler.invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Throws:

   - **YRException** - Unified exception types thrown.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.
