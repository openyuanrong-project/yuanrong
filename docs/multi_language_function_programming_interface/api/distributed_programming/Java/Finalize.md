# Finalize

## YR.Finalize()

package: `org.yuanrong.api`.

### public class YR extends YRCall

#### Interface description

##### public static synchronized void finalize(String ctx) throws YRException

The openYuanrong termination interface releases resources, including created function instances, to prevent resource leaks.

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

- Parameters:

   - **ctx** - Optional parameter: The tenant context to be terminated (can be obtained from the return value of [YR.init](init.md) .

- Throws:

   - **YRException** - Calling finalize before init will throw the exception.

##### public static void Finalize() throws YRException

The openYuanrong termination interface releases resources, including created function instances, to prevent resource leaks.

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

// Instance example
InstanceHandler counter = YR.instance(Counter::new).invoke();
ObjectRef objectRef = counter.function(Counter::increment).invoke();
System.out.println(YR.get(objectRef, 3000));

// Function example
ObjectRef res = YR.function(MyYRApp::myFunction).invoke();
System.out.println(YR.get(res, 3000));
ObjectRef objRef2 = YR.function(MyYRApp::functionWithAnArgument).invoke(1);
System.out.println(YR.get(objRef2, 3000));
YR.Finalize();
```

- Throws:

   - **YRException** - Calling Finalize before init will throw the exception.
