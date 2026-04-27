# setUrn

包名: `org.yuanrong.call`。

## 接口说明

### public CppFunctionHandler&lt;R&gt; setUrn(String urn)

java 调用 cpp 无状态函数时，为函数设置 functionUrn。

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    CppFunctionHandler&lt;R&gt;，内置 invoke 方法，可对该 cpp 函数实例进行创建并调用。

### public CppInstanceCreator setUrn(String urn)

java 调用 cpp 有状态函数时，为函数设置 functionUrn。

```java

CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter","FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke(1);
ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(5);
int res = (int)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    [CppInstanceCreator](CppInstanceCreator.md)，内置 invoke 方法，可对该 cpp 函数类实例进行创建。

### public JavaFunctionHandler&lt;R&gt; setUrn(String urn)

java 调用 java 无状态函数时，为函数设置 functionUrn。

```java

ObjectRef ref1 = YR.function(JavaFunction.of("com.example.YrlibHandler$MyYRApp", "smallCall", String.class))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
String res = (String)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    JavaFunctionHandler&lt;R&gt;，内置 invoke 方法，可对该 java 函数实例进行创建并调用。

### public JavaInstanceCreator setUrn(String urn)

 java 调用 java 有状态函数时，为函数设置 functionUrn。

```java

JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp"))
    .setUrn("sn:cn:yrk:default:function:0-perf-callee:$latest").invoke();
ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("smallCall", String.class)).invoke();
String res = (String)YR.get(ref1, 100);
```

- 参数：

   - **urn** - 函数 urn，可在函数部署之后获取。

- 返回：

    [JavaInstanceCreator](JavaInstanceCreator.md)，内置 invoke 方法，可对该 java 函数类实例进行创建。
