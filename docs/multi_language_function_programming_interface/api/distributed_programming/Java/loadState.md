# loadState

package: `package org.yuanrong.api`.

## loadState()

:::{Note}

Constraints:

- The Java language does not currently support local mode.
- Cluster mode is only supported for use in remote code.
- If `loadState` cannot retrieve the saved instance state, it will throw a `Failed to load state` exception.
- The class code corresponding to the instance must strictly adhere to the JavaBeans specification. For a property named `foo`, the getter method should be named `getFoo` or `isFoo` (for boolean types), and the setter method should be named `setFoo`. Otherwise, a `Catch MismatchedInputException while running LoadInstance` exception will be thrown when deserializing the class instance.

:::

### Interface description

#### public static void loadState() throws YRException

Used to load the saved instance state.

```java
public class Counter {
    private int cnt = 0;

    public int addOne() {
        this.cnt += 1;
        return this.cnt;
    }

    public int getCnt() {
        return this.cnt;
    }

    public void save() throws YRException {
        YR.saveState();
    }

    public boolean load() throws YRException {
        YR.loadState();
        return true;
    }
}

public static class MyYRApp{
    public static void main(String[] args) throws Exception {
        Config conf = new Config("FunctionURN", "ip", "ip", "", false);
        YR.init(conf);
        InstanceHandler ins = YR.instance(Counter::new).invoke();
        ins.function(Counter::save).invoke();

        ObjectRef ref = ins.function(Counter::addOne).invoke();
        Assert.assertEquals(YR.get(ref, 10), 1);

        ref = ins.function(Counter::load).invoke();
        YR.get(ref, 10);

        ref = ins.function(Counter::get).invoke();
        Assert.assertEquals(YR.get(ref, 10), 0);
        YR.Finalize();
    }
}
```

- Throws:

   - **YRException** - If obtaining the instance state fails, a `Failed to load state` exception will be thrown.

#### public static void loadState(int timeoutSec) throws YRException

Used to load the saved instance state.

- Parameters:

   - **timeoutSec** (int) – The timeout, in seconds, defaults to ``30``.

- Throws:

   - **YRException** - If obtaining the instance state fails, a `Failed to load state` exception will be thrown.
