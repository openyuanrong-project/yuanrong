# InstanceHandler function

package: `org.yuanrong.call`.

## public class InstanceHandler

Create an operation class for creating a Java stateful function instance.

:::{Note}

The class [InstanceHandler](InstanceHandler.md) is the handle returned after creating a Java class instance; it is the return type of the interface [InstanceCreator.invoke](InstanceCreator.md) after creating a Java class instance.

Users can use the function method of the built-in[InstanceHandler](InstanceHandler.md) to create a Java class instance member method handle and return the handle class [InstanceFunctionHandler](InstanceFunctionHandler.md).

:::

### Interface description

#### public VoidInstanceFunctionHandler function(YRFuncVoid0 func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 0 input parameters and 0 return values.

- Parameters:

   - **func** - YRFuncVoid0 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public &lt;T0&gt; VoidInstanceFunctionHandler function(YRFuncVoid1&lt;T0&gt; func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 1 input parameter and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **func** - YRFuncVoid1 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1> VoidInstanceFunctionHandler function(YRFuncVoid2<T0, T1> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 2 input parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **func** - YRFuncVoid2 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2> VoidInstanceFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 3 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **func** - YRFuncVoid3 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T4> VoidInstanceFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 4 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **func** - YRFuncVoid4 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T4, T5> VoidInstanceFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 5 parameters and 0 return values.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;T5&gt;** - Input parameter type.
   - **func** - YRFuncVoid5 Class instance.

- Returns:

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) Instance.

#### public &lt;R&gt; InstanceFunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 0 input parameters and 1 return value.

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc0 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc1<T0, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 1 input parameter and 1 return value.

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
ObjectRef ref = InstanceHandler.function(Counter::Add).invoke(5);
int res = (int)YR.get(ref, 100);
InstanceHandler.terminate();
```

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc1 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc2<T0, T1, R> func)

Java functions in the cloud can call member functions of Java class instances in the cloud, supporting user functions with 2 input parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc2 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc3<T0, T1, T2, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 3 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc3 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T3, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc4<T0, T1, T2, T3, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 4 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc4 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.

#### public <T0, T1, T2, T3, T4, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc5<T0, T1, T2, T3, T4, R> func)

Java functions under the cloud call member functions of Java class instances on the cloud, supporting user functions with 5 parameters and 1 return value.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** – Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - YRFunc5 Class instance.

- Returns:

    [InstanceFunctionHandler](InstanceFunctionHandler.md) Instance.
