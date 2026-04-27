# Instance

## YR.instance()

package: `org.yuanrong`.

Under the cloud, a Java function creates a Java class instance on the cloud and calls member functions. The current constructor supports 0 to 5 parameters.
If a name is passed, or both a name and a namespace are passed, it indicates that the instance is a named instance.
The instance name is specified by the user and can be reused through the instance name.

### public class YRCall extends YRGetInstance

The type YR Call.

#### Interface description

##### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func)

Instance instance creator.

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

- Parameters:

   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

##### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func, String name)

Instance instance creator.

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

- Parameters:

   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static &lt;A&gt; InstanceCreator&lt;A&gt; instance(YRFunc0&lt;A&gt; func, String name, String nameSpace)

Instance instance creator.

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

- Parameters:

   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func, String name)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, A> InstanceCreator&lt;A&gt; instance(YRFunc1<T0, A> func, String name, String nameSpace)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, A> InstanceCreator&lt;A&gt; instance(YRFunc2<T0, T1, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func, String name)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, A> InstanceCreator &lt;A&gt; instance (YRFunc2<T0, T1, A> func, String name, String nameSpace)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, A> InstanceCreator &lt;A&gt; instance (YRFunc3<T0, T1, T2, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, A> InstanceCreator &lt;A&gt; instance (YRFunc3<T0, T1, T2, A > func, String name)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, A> InstanceCreator< A> instance (YRFunc3<T0, T1, T2, A> func, String name, String nameSpace)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func, String name)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, A> InstanceCreator< A> instance (YRFunc4<T0, T1, T2, T3, A> func, String name, String nameSpace)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func, String name)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    InstanceCreator Instance.

#### public static <T0, T1, T2, T3, T4, A> InstanceCreator< A> instance (YRFunc5<T0, T1, T2, T3, T4, A> func, String name, String nameSpace)

Instance instance creator.

- Parameters:

   - **&lt;T0&gt;** - Input parameter type.
   - **&lt;T1&gt;** - Input parameter type.
   - **&lt;T2&gt;** - Input parameter type.
   - **&lt;T3&gt;** - Input parameter type.
   - **&lt;T4&gt;** - Input parameter type.
   - **&lt;A&gt;** - Return value type.
   - **func** - Function name.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    InstanceCreator Instance.

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass)

Instance java instance creator.

- Parameters:

   - **javaInstanceClass** - the java instance class.

- Returns:

    JavaInstanceCreator Instance.

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name)

Instance java instance creator.

- Parameters:

   - **javaInstanceClass** - the java instance class.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    JavaInstanceCreator Instance.

#### public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name, String nameSpace)

Instance java instance creator.

- Parameters:

   - **javaInstanceClass** - the java instance class.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    JavaInstanceCreator Instance.

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass)

Instance cpp instance creator.

- Parameters:

   - **cppInstanceClass** - the cpp instance class.

- Returns:

    CppInstanceCreator Instance.

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name)

Instance cpp instance creator.

- Parameters:

   - **cppInstanceClass** - the cpp instance class.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.

- Returns:

    CppInstanceCreator Instance.

#### public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name, String nameSpace)

Instance cpp instance creator.

- Parameters:

   - **cppInstanceClass** - the cpp instance class.
   - **name** - The instance name of the named instance, the second parameter. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name is concatenated into nameSpace-name. This field is currently used only for concatenation.

- Returns:

    CppInstanceCreator Instance.
