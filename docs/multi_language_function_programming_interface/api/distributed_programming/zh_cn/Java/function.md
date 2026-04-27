# function

## YR.function()

包名: `org.yuanrong`.

云下 Java 函数调用云上 Java 函数，openYuanrong 当前支持 0 ~ 5 个入参，一个返回值的用户函数。

### public class YRCall extends YRGetInstance

YR Call 类型。

#### 接口说明

##### public static &lt;R&gt; FunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

创建 FunctionHandler 实例。

```java

public static class MyYRApp {
   public static String myFunction() {
      return "hello world";
   }
}

public static void main(String[] args) throws YRException {
   Config conf = new Config("FunctionURN", "ip", "ip", "", false);
   YR.init(conf);
   FunctionHandler<String> f_h = YR.function(MyYRApp::myFunction);
   ObjectRef res = f_h.invoke();
   System.out.println("myFunction invoke ref:" + res.getObjId());
   YR.Finalize();
}
```

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static <T0, R> FunctionHandler&lt;R&gt; function(YRFunc1<T0, R> func)

创建 FunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static <T0, T1, R> FunctionHandler&lt;R&gt; function(YRFunc2<T0, T1, R> func)

创建 FunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static <T0, T1, T2, R> FunctionHandler&lt;R&gt; function(YRFunc3<T0, T1, T2, R> func)

创建 FunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static <T0, T1, T2, T3, R> FunctionHandler&lt;R&gt; function(YRFunc4<T0, T1, T2, T3, R> func)

创建 FunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static <T0, T1, T2, T3, T4, R> FunctionHandler&lt;R&gt; function(YRFunc5<T0, T1, T2, T3, T4, R> func)

创建 FunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;R&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    FunctionHandler 实例。

##### public static VoidFunctionHandler function(YRFuncVoid0 func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **func** - 函数名。

- 返回：

   VoidFunctionHandler 实例。

##### public static &lt;T0&gt; VoidFunctionHandler function(YRFuncVoid1&lt;T0&gt; func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **func** - 函数名。

- 返回：

    VoidFunctionHandler 实例。

##### public static <T0, T1> VoidFunctionHandler function(YRFuncVoid2<T0, T1> func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **func** - 函数名。

- 返回：

    VoidFunctionHandler 实例。

##### public static <T0, T1, T2> VoidFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **func** - 函数名。

- 返回：

    VoidFunctionHandler 实例。

##### public static <T0, T1, T2, T4> VoidFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **func** - 函数名。

- 返回：

    VoidFunctionHandler 实例。

##### public static <T0, T1, T2, T4, T5> VoidFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

创建 VoidFunctionHandler 实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;T5&gt;** - 入参类型。
   - **func** - 函数名。

- 返回：

    VoidFunctionHandler 实例。

##### public static &lt;R&gt; JavaFunctionHandler&lt;R&gt; function(JavaFunction&lt;R&gt; javaFunction)

创建 JavaFunctionHandler 实例。

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **javaFunction** -  java 函数名。

- 返回：

   JavaFunctionHandler 实例。

##### public static &lt;R&gt; CppFunctionHandler&lt;R&gt; function(CppFunction&lt;R&gt; cppFunction)

创建 CppFunctionHandler 实例。

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **cppFunction** -  C++ 函数名。

- 返回：

   CppFunctionHandler 实例。
