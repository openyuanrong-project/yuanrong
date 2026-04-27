# Instance

## YR.instance()

包名: `org.yuanrong`。

云下 Java 函数创建云上 Java 类实例并调用成员函数，当前构造函数支持 0 ~ 5 个入参。若传入 name，或者同时传入 name 和 nameSpace，表示该实例为具名实例，实例名由用户指定，可以通过该实例名进行实例的复用。

### public class YRCall extends YRGetInstance

YR Call 类型。

#### 接口说明

##### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

```java

public static class MyYRApp{
    public static String myFunction() throws YRException, ExecutionException{
        return "hello world";
    }
}
public static void main(String[] args) throws Exception {
    Config conf = new Config("FunctionURN", "ip", "ip", "", false);
    YR.init(conf);
    MyYRApp myapp = new MyYRApp();
    InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new);
    InstanceHandler h_myYRapp = myYRapp.invoke();
    System.out.println("InstanceHandler:" + h_myYRapp);
    InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
    ObjectRef res = f_h_myYRapp.invoke();
    System.out.println("myFunction invoke ref:" + res.getObjId());
    YR.Finalize();
}
```

- 参数：

   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

##### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

```java

public static class MyYRApp{
     public static String myFunction() throws YRException, ExecutionException{
         return "hello world";
     }
 }
 public static void main(String[] args) throws Exception {
     Config conf = new Config("FunctionURN", "ip", "ip", "", false);
     YR.init(conf);
     MyYRApp myapp = new MyYRApp();
     // The instance name of this named instance is funcB
     InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new, "funcB");
     InstanceHandler h_myYRapp = myYRapp.invoke();
     System.out.println("InstanceHandler:" + h_myYRapp);
     InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
     ObjectRef res = f_h_myYRapp.invoke();
     System.out.println("myFunction invoke ref:" + res.getObjId());
     // A handle to a named instance of funcB will be generated, and the funcB instance will be reused.
     InstanceCreator<MyYRApp> creator = YR.instance(MyYRApp::new, "funcB");
     InstanceHandler handler = creator.invoke();
     System.out.println("InstanceHandler:" + handler);
     InstanceFunctionHandler<String> funcHandler = handler.function(MyYRApp::myFunction);
     res = funcHandler.invoke();
     System.out.println("myFunction invoke ref:" + res.getObjId());
     YR.Finalize();
 }
```

- 参数：

   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

```java

public static class MyYRApp{
     public static String myFunction() throws YRException, ExecutionException{
         return "hello world";
     }
 }
 public static void main(String[] args) throws Exception {
     Config conf = new Config("FunctionURN", "ip", "ip", "", false);
     YR.init(conf);
     MyYRApp myapp = new MyYRApp();
     // The instance name of this named instance is nsA-funcB
     InstanceCreator<MyYRApp> myYRapp = YR.instance(MyYRApp::new, "funcB", "nsA");
     InstanceHandler h_myYRapp = myYRapp.invoke();
     System.out.println("InstanceHandler:" + h_myYRapp);
     InstanceFunctionHandler<String> f_h_myYRapp = h_myYRapp.function(MyYRApp::myFunction);
     ObjectRef res = f_h_myYRapp.invoke();
     System.out.println("myFunction invoke ref:" + res.getObjId());
     // A handle to the named instance of nsA-funcB will be generated, reusing the nsA-funcB instance
     InstanceCreator<MyYRApp> creator = YR.instance(MyYRApp::new, "funcB", "nsA");
     InstanceHandler handler = creator.invoke();
     System.out.println("InstanceHandler:" + handler);
     InstanceFunctionHandler<String> funcHandler = handler.function(MyYRApp::myFunction);
     res = funcHandler.invoke();
     System.out.println("myFunction invoke ref:" + res.getObjId());
     YR.Finalize();
 }
```

- 参数：

   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, A> InstanceCreator &lt;A&gt; instance (YRFunc3<T0, T1, T2, A> func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, A> InstanceCreator &lt;A&gt; instance (YRFunc3<T0, T1, T2, A > func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, A> InstanceCreator< A> instance (YRFunc3<T0, T1, T2, A> func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func, String name)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func, String name, String nameSpace)

创建 [InstanceCreator](InstanceCreator.md) 类实例。

- 参数：

   - **&lt;T0&gt;** - 入参类型。
   - **&lt;T1&gt;** - 入参类型。
   - **&lt;T2&gt;** - 入参类型。
   - **&lt;T3&gt;** - 入参类型。
   - **&lt;T4&gt;** - 入参类型。
   - **&lt;A&gt;** - 返回值类型。
   - **func** - 函数名。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [InstanceCreator](InstanceCreator.md) 类实例。

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass)

创建 [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

- 参数：

   - **javaInstanceClass** - java 类。

- 返回：

    [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name)

创建 [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

- 参数：

   - **javaInstanceClass** - java 类。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name, String nameSpace)

创建 [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

- 参数：

   - **javaInstanceClass** - java 类。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [JavaInstanceCreator](JavaInstanceCreator.md) 类实例。

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass)

创建 [CppInstanceCreator](CppInstanceCreator.md) 类实例。

- 参数：

   - **cppInstanceClass** - cpp 类。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md) 类实例。

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name)

创建 [CppInstanceCreator](CppInstanceCreator.md) 类实例。

- 参数：

   - **cppInstanceClass** - cpp 类。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md) 类实例。

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name, String nameSpace)

创建 [CppInstanceCreator](CppInstanceCreator.md) 类实例。

- 参数：

   - **cppInstanceClass** - cpp 类。
   - **name** - 具名实例的实例名，第二入参。当仅存在 name 时，实例名将设置成 name。
   - **nameSpace** - nameSpace 具名实例的命名空间，当 name 和 nameSpace 均存在时，实例名将拼接成 nameSpace-name，该字段目前仅用做拼接。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md) 类实例。
