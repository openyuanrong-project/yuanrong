# function

## YR.function()

package: `org.yuanrong`.

Under the cloud, Java function calls Java functions on the cloud. openYuanrong currently supports user functions with 0 to 5 input parameters and one return value.

### public class YRCall extends YRGetInstance

The type YR Call.

#### Interface description

##### public static &lt;R&gt; FunctionHandler&lt;R&gt; function(YRFunc0&lt;R&gt; func)

Function function handler.

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

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static <T0, R> FunctionHandler&lt;R&gt; function(YRFunc1<T0, R> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static <T0, T1, R> FunctionHandler&lt;R&gt; function(YRFunc2<T0, T1, R> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static <T0, T1, T2, R> FunctionHandler&lt;R&gt; function(YRFunc3<T0, T1, T2, R> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static <T0, T1, T2, T3, R> FunctionHandler&lt;R&gt; function(YRFunc4<T0, T1, T2, T3, R> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static <T0, T1, T2, T3, T4, R> FunctionHandler&lt;R&gt; function(YRFunc5<T0, T1, T2, T3, T4, R> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;R&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    FunctionHandler Instance.

##### public static VoidFunctionHandler function(YRFuncVoid0 func)

Function void function handler.

- Parameters:

   - **func** - Function name.

- Returns:

   VoidFunctionHandler Instance.

##### public static &lt;T0&gt; VoidFunctionHandler function(YRFuncVoid1&lt;T0&gt; func)

Function void function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **func** - Function name.

- Returns:

    VoidFunctionHandler Instance.

##### public static <T0, T1> VoidFunctionHandler function(YRFuncVoid2<T0, T1> func)

Function void function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **func** - Function name.

- Returns:

    VoidFunctionHandler Instance.

##### public static <T0, T1, T2> VoidFunctionHandler function(YRFuncVoid3<T0, T1, T2> func)

Function void function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **func** - Function name.

- Returns:

    VoidFunctionHandler Instance.

##### public static <T0, T1, T2, T4> VoidFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func)

Function void function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **func** - Function name.

- Returns:

    VoidFunctionHandler Instance.

##### public static <T0, T1, T2, T4, T5> VoidFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func)

Function function handler.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;T5&gt;** - Input parameter type.
   - **func** - Function name.

- Returns:

    VoidFunctionHandler Instance.

##### public static &lt;R&gt; JavaFunctionHandler&lt;R&gt; function(JavaFunction&lt;R&gt; javaFunction)

Function java function handler.

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **javaFunction** -  java Function name.

- Returns:

   JavaFunctionHandler Instance.

##### public static &lt;R&gt; CppFunctionHandler&lt;R&gt; function(CppFunction&lt;R&gt; cppFunction)

Function cpp function handler.

- Parameters:

   - **&lt;R&gt;** - Return value type.
   - **cppFunction** -  C++ Function name.

- Returns:

   CppFunctionHandler Instance.
