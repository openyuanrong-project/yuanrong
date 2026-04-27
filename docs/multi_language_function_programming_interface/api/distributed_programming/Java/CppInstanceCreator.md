# CppInstanceCreator

package: `org.yuanrong.call`.

## public class CppInstanceCreator

Create an operation class for creating a cpp stateful function instance.

:::{Note}

The CppInstanceCreator class is a Java function that creates C++ class instances; it is the return type of the interface [YR.instance(CppInstanceClass cppInstanceClass)](Instance.md).

Users can use the invoke method of CppInstanceCreator to create a C++ class instance and return the handle class CppInstanceHandler.

:::

### Interface description

#### public CppInstanceCreator(CppInstanceClass cppInstanceClass)

The constructor of CppInstanceCreator.

- Parameters:

   - **cppInstanceClass** - CppInstanceClass class instance.

#### public CppInstanceCreator(CppInstanceClass cppInstanceClass, String name, String nameSpace)

The constructor of CppInstanceCreator.

- Parameters:

   - **cppInstanceClass** - CppInstanceClass class instance.
   - **name** - The instance name of a named instance. When only name exists, the instance name will be set to name.
   - **nameSpace** -  Namespace of the named instance. When both name and nameSpace exist, the instance name will be concatenated into nameSpace-name. This field is currently only used for concatenation, and namespace isolation and other related functions will be completed later.

#### public CppInstanceHandler invoke(Object... args) throws YRException

The member method of the CppInstanceHandler class is used to create a cpp class instance.

- Parameters:

   - **args** - The input parameters required to create a class instance.

- Returns:

    [CppInstanceHandler](CppInstanceHandler.md) Class handle.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public CppInstanceCreator setUrn(String urn)

When Java calls a stateful function in C++, set the functionUrn for the function.

```java

CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter","FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    CppInstanceCreator, with built-in invoke method, can create instances of this cpp function class.

#### public CppInstanceCreator options(InvokeOptions opt)

The member method of the CppInstanceCreator class is used to dynamically modify the parameters for creating a C++ function instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
CppInstanceCreator cppInstanceCreator = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").options(invokeOptions);
CppInstanceHandler cppInstanceHandler = cppInstanceCreator.invoke(1);
ObjectRef ref = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    CppInstanceCreator Class handle.
