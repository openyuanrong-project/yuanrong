# CppFunctionHandler

package: `org.yuanrong.call`.

## public class CppFunctionHandler&lt;R>

Class for creating instances of stateless functions in C++.

:::{Note}

The CppFunctionHandler class is a handle for a C++ function created on the cloud by a Java function. It is the return type of the interface [YR.function(CppFunction&lt;R&gt; func)](function.md).

Users can use CppFunctionHandler to create and invoke C++ function instances.

:::

### Interface description

#### public CppFunctionHandler(CppFunction&lt;R&gt; func)

CppFunctionHandler constructor.

- Parameters:

   - **func** - CppFunction class instance.

#### public ObjectRef invoke(Object... args) throws YRException

Cpp function call interface.

```java

CppFunctionHandler cppFuncHandler = YR.function(CppFunction.of("Add", int.class));
ObjectRef ref = cppFuncHandler.invoke(1);
int result = YR.get(ref, 15);
```

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public CppFunctionHandler&lt;R&gt; setUrn(String urn)

When Java calls a stateless function in C++, set the functionUrn for the function.

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    CppFunctionHandler&lt;R&gt;, with built-in invoke method, can create and call the cpp function instance.

#### public CppFunctionHandler&lt;R&gt; options(InvokeOptions opt)

Used to dynamically modify the parameters of the called function.

- Parameters:

   - **opt** - See [InvokeOptions](InvokeOptions.md) for details.

- Returns:

    CppFunctionHandler&lt;R&gt;, Handles processed by C++ functions.
