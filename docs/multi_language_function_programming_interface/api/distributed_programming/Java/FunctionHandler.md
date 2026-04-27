# FunctionHandler

package: `org.yuanrong.call`.

## public class FunctionHandler<R> extends Handler

Create an operation class for creating a stateless Java function instance.

:::{Note}

The FunctionHandler class is the handle for Java functions on the cloud; it is the return type of the interface [YR.function(YRFuncR&lt;R&gt; func)](function.md).

Users can use FunctionHandler to create and invoke Java function instances.

:::

### Interface description

#### public FunctionHandler(YRFuncR&lt;R&gt; func)

FunctionHandler Constructor.

- Parameters:

   - **func** - Java function name, openYuanrong currently supports 0 ~ 5 parameters and one return value for user functions.

#### public ObjectRef invoke(Object... args) throws YRException

Java function call interface.

- Parameters:

   - **args** - The input parameters required to call the specified method.

- Returns:

    ObjectRef: The “id” of the method’s return value in the data system. Use [YR.get()](get.md) to get the actual return value of the method.

- Throws:

   - **YRException** - Unified exception types thrown.

#### public FunctionHandler&lt;R&gt; options(InvokeOptions opt)

Used to dynamically modify the parameters of the called function.

```java

Config conf = new Config("FunctionURN", "ip", "ip", "", false);
YR.init(conf);
InvokeOptions opts = new InvokeOptions();
opts.setCpu(300);
FunctionHandler<String> f_h = YR.function(MyYRApp::myFunction).options(opts);
ObjectRef res = f_h.invoke();
System.out.println("myFunction invoke ref:" + res.getObjId());
YR.Finalize();
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    FunctionHandler Class handle.
