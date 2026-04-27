# setUrn

package: `org.yuanrong.call`.

## Interface description

### public CppFunctionHandler&lt;R&gt; setUrn(String urn)

When Java calls a stateless function in C++, set the functionUrn for the function.

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    CppFunctionHandler&lt;R&gt;, with built-in invoke method, can create and call the cpp function instance.

### public CppInstanceCreator setUrn(String urn)

When Java calls a stateful function in C++, set the functionUrn for the function.

```java

CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter","FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    [CppInstanceCreator]struct-CppInstanceCreator.md), with built-in invoke method, can create instances of this cpp function class.

### public JavaFunctionHandler&lt;R&gt; setUrn(String urn)

When Java calls a stateless function in Java, set the functionUrn for the function.

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
String res = (String)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    JavaFunctionHandler&lt;R&gt;, with built-in invoke method, can create and invoke the java function instance.

### public JavaInstanceCreator setUrn(String urn)

When Java calls a Java stateful function, set the functionUrn for the function.

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp"))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref1, 100);
```

- Parameters:

   - **urn** - functionUrn, can be obtained after the function is deployed.

- Returns:

    [JavaInstanceCreator](JavaInstanceCreator.md), with built-in invoke method, can create instances of this Java function class.
