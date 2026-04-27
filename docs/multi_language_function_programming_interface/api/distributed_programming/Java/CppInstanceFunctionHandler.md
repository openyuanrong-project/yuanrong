# CppInstanceFunctionHandler

package: `org.yuanrong.call`.

## public class CppInstanceFunctionHandler&lt;R>

Operation class that calls a C++ stateful function instance member function.

:::{Note}

The CppInstanceFunctionHandler class is the handle of the member function of the C++ class instance after the Java function creates the C++ class instance. It is the return value type of the interface [CppInstanceHandler.function](CppInstanceHandler.md).

Users can use the invoke method of CppInstanceFunctionHandler to call member functions of C++ class instances.

:::

- Parameters:

   - **&lt;R&gt;** - The specific type of a member function.

### Interface description

#### public ObjectRef invoke(Object... args) throws YRException

The member method of the CppInstanceFunctionHandler class is used to call the member function of a cpp class instance.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter","FactoryCreate")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
CppInstanceFunctionHandler cppInsFuncHandler = cppInstanceHandler.function(CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInsFuncHandler.invoke(5);
int res = (int)YR.get(ref, 100);
```

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public CppInstanceFunctionHandler&lt;R&gt; options(InvokeOptions opt)

Member method of the CppInstanceFunctionHandler class, used to dynamically modify the parameters of the called function.

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    CppInstanceFunctionHandler Class handle.

#### CppInstanceFunctionHandler(String instanceId, String functionId, String className, CppInstanceMethod&lt;R&gt; cppInstanceMethod)

The constructor of CppInstanceFunctionHandler.

- Parameters:

   - **instanceId** - cpp function instance ID.
   - **functionId** - cpp function deployment returns ID.
   - **className** - cpp function class name.
   - **cppInstanceMethod** - cppInstanceMethod class instance.
