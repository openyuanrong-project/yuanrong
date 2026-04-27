# InstanceCreator

package: `org.yuanrong.call`.

## public class InstanceCreator<A> extends Handler

Creates an operation class for creating a Java stateful function instance.

:::{Note}

The instanceCreator is the creator of Java class instances; it is the return type of the interface [YR.instance(YRFuncR&lt;A&gt; func)](Instance.md).

Users can use the invoke method of instanceCreator to create Java class instances and return the [InstanceHandler](InstanceHandler.md) class handle.

:::

### Interface description

#### public InstanceCreator(YRFuncR&lt;A&gt; func)

The constructor of InstanceCreator.

- Parameters:

   - **func** - YRFuncR class instance, Java function name.

#### public InstanceCreator(YRFuncR&lt;A&gt; func, ApiType apiType)

The constructor of InstanceCreator.

- Parameters:

   - **func** - YRFuncR class instance, Java function name.
   - **apiType** - The enumeration class has two values: Function and Posix. It is used internally by openYuanrong to distinguish function types. The default is Function.

#### public InstanceCreator(YRFuncR&lt;A&gt; func, String name, String nameSpace)

The constructor of InstanceCreator.

- Parameters:

   - **func** - YRFuncR class instance, Java function name.
   - **name** - The instance name of a named instance. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name will be concatenated into nameSpace-name. This field is currently only used for concatenation, and namespace isolation and other related functions will be completed later.

#### public InstanceCreator(YRFuncR&lt;A&gt; func, String name, String nameSpace, ApiType apiType)

The constructor of InstanceCreator.

- Parameters:

   - **func** - YRFuncR class instance, Java function name.
   - **name** - The instance name of a named instance. When only name exists, the instance name will be set to name.
   - **nameSpace** - Namespace of the named instance. When both name and nameSpace exist, the instance name will be concatenated into nameSpace-name. This field is currently only used for concatenation, and namespace isolation and other related functions will be completed later.
   - **apiType** - The enumeration class has two values: Function and Posix. It is used internally by openYuanrong to distinguish function types. The default is Function.

#### public InstanceHandler invoke(Object... args) throws YRException

The member method of the InstanceCreator class is used to create a Java class instance.

- Parameters:

   - **args** - The input parameters required to create a class instance.

- Returns:

    [InstanceHandler](InstanceHandler.md) Class handle.

- Throws：

   - **YRException** - Unified exception types thrown.

#### public InstanceCreator&lt;A&gt; options(InvokeOptions opt)

The member method of the InstanceCreator class is used to dynamically modify the parameters for creating a Java function instance.

```java

InvokeOptions invokeOptions = new InvokeOptions();
invokeOptions.setCpu(1500);
invokeOptions.setMemory(1500);
InstanceCreator instanceCreator = YR.instance(Counter::new).options(invokeOptions);
InstanceHandler instanceHandler = instanceCreator.invoke(1);
```

- Parameters:

   - **opt** - Function call options, used to specify functions such as calling resources.

- Returns:

    InstanceCreator Class handle.
