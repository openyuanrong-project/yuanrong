# saveState

包名：`package org.yuanrong.api`。

## saveState()

:::{Note}

约束：

- java 语言暂不支持本地模式。
- 集群模式仅支持在远端代码中使用。
- 保存的实例状态数据最大限制为 4M，超过最大限制时，使用 saveState 保存状态会失败，抛出`Failed to save state`异常。
- 实例对应的类代码需要严格遵守 Java Beans 规范。属性名为 foo 的 getter 方法应命名为 getFoo 或者 isFoo（对于布尔类型）, setter 方法应命名为 setFoo。否则反序列化类实例时会抛出`Catch MismatchedInputException while running LoadInstance`异常。

:::

### 接口说明

#### public static void saveState() throws YRException

用于保存实例状态。

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

- 抛出：

   - **YRException** - 保存状态会失败抛出 `Failed to save state` 异常。

#### public static void saveState(int timeoutSec) throws YRException

用于保存实例状态。

- 参数：

   - **timeoutSec** (int) – 超时时间，单位为秒，默认 ``30``。

- 抛出：

   - **YRException** - 保存状态会失败抛出 `Failed to save state` 异常。
