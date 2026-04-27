# InstanceHandler function

包名：`org.yuanrong.call`.

## public class InstanceHandler

创建 java 有状态函数类实例的操作类。

:::{Note}

类 [InstanceHandler](InstanceHandler.md) 是创建 Java 类实例后返回的句柄。是接口 [InstanceCreator.invoke](InstanceCreator.md) 创建 Java 类实例后的返回值类型。

用户可以使用 [InstanceHandler](InstanceHandler.md) 内置的 function 方法创建 Java 类实例成员方法句柄，并返回句柄类 [InstanceFunctionHandler](InstanceFunctionHandler.md)。

:::

### 接口说明

#### public VoidInstanceFunctionHandler function(YRFuncVoid0 func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 0 个入参，0 个返回值的用户函数。

- 参数：

   - **func** - YRFuncVoid0 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public &lt;T0&gt; VoidInstanceFunctionHandler function(YRFuncVoid1&lt;T0&gt; func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 1 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **func** - YRFuncVoid1 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1> VoidInstanceFunctionHandler function(YRFuncVoid2<T0, T1> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 2 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **func** - YRFuncVoid2 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2> VoidInstanceFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 3 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **func** - YRFuncVoid3 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T4> VoidInstanceFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 4 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **func** - YRFuncVoid4 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T4, T5> VoidInstanceFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 5 个入参，0 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;T5&gt;** - 入参类型。
   - **func** - YRFuncVoid5 类实例。

- 返回：

    [VoidInstanceFunctionHandler](VoidInstanceFunctionHandler.md) 实例。

#### public &lt;R&gt; InstanceFunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 0 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc0 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc1<T0, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 1 个入参，1 个返回值的用户函数。

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
ObjectRef ref = InstanceHandler.function(Counter::Add).invoke(5);
int res = (int)YR.get(ref, 100);
InstanceHandler.terminate();
```

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc1 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc2<T0, T1, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 2 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc2 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc3<T0, T1, T2, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 3 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc3 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T3, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc4<T0, T1, T2, T3, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 4 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc4 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。

#### public <T0, T1, T2, T3, T4, R> InstanceFunctionHandler&lt;R&gt; function(YRFunc5<T0, T1, T2, T3, T4, R> func)

云下 Java 函数调用云上 Java 类实例的成员函数，支持 5 个入参，1 个返回值的用户函数。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** – 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - YRFunc5 类实例。

- 返回：

    [InstanceFunctionHandler](InstanceFunctionHandler.md) 实例。
