# CppInstanceHandler function

package: `org.yuanrong.call`.

## public class CppInstanceHandler

Create an operation class for creating a cpp stateful function instance.

:::{Note}

The [CppInstanceHandler](CppInstanceHandler.md) class is the handle returned after a Java function creates a cpp class instance; it is the return type of the [CppInstanceCreator.invoke](CppInstanceCreator.md) interface after creating a cpp class instance.

Users can use the function method of [CppInstanceHandler](CppInstanceHandler.md) to create a cpp class instance member method handle and return a handle class [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md).

:::

### Interface description

#### public &lt;R&gt; CppInstanceFunctionHandler&lt;R&gt; function(CppInstanceMethod&lt;R&gt; cppInstanceMethod)

The member method of the [CppInstanceHandler](CppInstanceHandler.md) class is used to return the handle of the member function of the cloud C++ class instance.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
CppInstanceFunctionHandler cppInstFuncHandler = cppInstanceHandler.function(
    CppInstanceMethod.of("Add", int.class));
ObjectRef ref = cppInstFuncHandler.invoke(5);
int res = (int) YR.get(ref, 100);
```

- Parameters:

   - **&lt;R&gt;** - the type parameter.
   - **cppInstanceMethod** - CppInstanceMethod class instance.

- Returns:

    [CppInstanceFunctionHandler](CppInstanceFunctionHandler.md) Instance.
