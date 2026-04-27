# yrShutdown

:::{Note}

约束：

- 用户自定义`yrShutdown`时，函数签名必须与接口原型一致。否则云上将无法找到对应函数执行优雅退出。

- `yrShutdown`执行时间如果超过 `gracePeriodSeconds` 秒，云上有状态实例不会等待函数执行完毕，会被立即回收。

:::

## 接口说明

### public void yrShutdown(int gracePeriodSeconds)

用户可以使用此接口执行数据清理或数据持久化等操作。

在云上执行的有状态函数中，增加名为`yrShutdown`的优雅退出函数。函数中支持调用 openYuanrong 接口。`yrShutdown`函数会在以下场景中执行：

1. 在 runtime 收到 shutdown 请求时（例如云下用户调用 terminate 接口）。

2. runtime 捕获到 Sigterm 信号时（例如集群执行缩容策略，runtime-manager 捕获到 Sigterm 信号后转发到 runtime）。

```java

// 有状态函数中定义 yrShutdown
public class Counter {
    private int cnt = 0;

    public int add() {
        this.cnt += 1;
        return this.cnt;
    }

    public void yrShutdown(int gracePeriodSeconds) throws YRException {
        SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
        YR.kv().set("myKey", "myValue".getBytes(StandardCharsets.UTF_8), setParam);
    }
}

public void main(String[] args) throws Exception {
    // 用例 1：未配置 GRACEFUL_SHUTDOWN_TIME，使用系统配置优雅退出时间
    YR.kv().del("myKey");
    InstanceHandler ins = YR.instance(Counter::new).invoke();
    ObjectRef ref = ins.function(Counter::add).invoke();
    YR.get(ref, 10);
    ins.terminate();
    // 云下使用 kv 接口获取 my_key 映射的值将为 'my_value'
    byte[] byteArray = YR.kv().get("myKey", 30);
    String shutdownValue = new String(byteArray, StandardCharsets.UTF_8);
    Assert.assertEquals(shutdownValue, "myValue");

    // 用例 2：配置 GRACEFUL_SHUTDOWN_TIME
    YR.kv().del("myKey");
    InvokeOptions option = new InvokeOptions();
    option.getCustomExtensions().put("GRACEFUL_SHUTDOWN_TIME","10");
    InstanceHandler ins2 = YR.instance(Counter::new).options(option).invoke();
    ObjectRef ref2 = ins2.function(Counter::add).invoke();
    YR.get(ref2, 10);
    ins2.terminate();
    // 云下使用 kv 接口获取 my_key 映射的值将为 'my_value'
    byte[] byteArray2 = YR.kv().get("myKey", 30);
    String shutdownValue2 = new String(byteArray2, StandardCharsets.UTF_8);
    Assert.assertEquals(shutdownValue2, "myValue");

    YR.Finalize();
}
```

- 参数：

   - **gracePeriodSecond** (int) – 函数执行超时时间，单位：秒，默认值：``60``。
