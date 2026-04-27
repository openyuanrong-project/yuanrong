# Finalize

## YR.Finalize()

包名： `org.yuanrong.api`。

### public class YR extends YRCall

#### 接口说明

##### public static synchronized void finalize(String ctx) throws YRException

openYuanrong 终止接口释放资源，包括创建的函数实例，以防止资源泄漏。

```java

try {
    ClientInfo info = YR.init(conf);
    String ctx = info.getContext();
    InstanceHandler instanceHandler = YR.instance(YrlibHandler.MyYRApp::new).invoke();
    ObjectRef ref1 = instanceHandler.function(MyYRApp::longCall).invoke();
    String res = (String)YR.get(ref1, 10000);
    System.out.println(res);
    YR.finalize(ctx);
} catch (YRException exp) {
    System.out.println(exp.getMessage());
}
```

- 参数：

   - **ctx** - 可选参数：要终止的租户上下文。（可以通过 [YR.init](init.md) 的返回值获得）。

- 抛出：

   - **YRException** - 在调用 init 之前调用 finalize，或者在租户上下文未初始化时传递租户上下文参数，都会抛出此异常。

##### public static void Finalize() throws YRException

openYuanrong 终止接口释放资源，包括创建的函数实例，以防止资源泄漏。

```java

public static class Counter {
    private int value = 0;

    public int increment() {
        this.value += 1;
        return this.value;
    }

    public static int functionWithAnArgument(int value) {
        return value + 1;
    }
}

public static class MyYRApp {
    public static int myFunction() {
        return 1;
    }

    public static int functionWithAnArgument(int value) {
        return value + 1;
    }
}
```

```java

// 实例样例
InstanceHandler counter = YR.instance(Counter::new).invoke();
ObjectRef objectRef = counter.function(Counter::increment).invoke();
System.out.println(YR.get(objectRef, 3000));

// 函数样例
ObjectRef res = YR.function(MyYRApp::myFunction).invoke();
System.out.println(YR.get(res, 3000));
ObjectRef objRef2 = YR.function(MyYRApp::functionWithAnArgument).invoke(1);
System.out.println(YR.get(objRef2, 3000));
YR.Finalize();
```

- 抛出：

   - **YRException** - 在调用 init 之前调用 Finalize，或者在租户上下文未初始化时传递租户上下文参数，都会抛出此异常。
