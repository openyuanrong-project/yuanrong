# yrRecover

:::{Note}

约束：用户自定义`yrRecover`时，函数签名必须与接口原型一致。否则云上将无法找到对应函数并执行。

:::

## 接口说明

### public void yrRecover()

用户可以使用此接口执行状态恢复、初始化状态等操作。
在云上执行的有状态函数中，增加名为 `yrRecover` 的钩子函数。函数中支持调用 openYuanrong 接口。`yrRecover` 函数会在以下场景中执行：

1. 在 runtime 收到 recover 请求时（例如云下用户在 InvokeOptions 中配置 recoverRetryTimes > 0，且云上实例因为异常退出）。

样例 1：

```java

// 有状态函数中定义 yrRecover
public class Counter {
    private int cnt = 0;

    public static long getPid () {
        String processName = ManagementFactory.getRuntimeMXBean().getName();
        long pid = Long.parseLong(processName.split("@")[0]);
        return pid;
    }

    public void yrRecover() throws YRException {
        SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
        YR.kv().set("myKey", "myValue".getBytes(StandardCharsets.UTF_8), setParam);
    }
}

public void main(String[] args) throws Exception {
    YR.kv().del("myKey");
    InstanceHandler ins = YR.instance(Counter::new).invoke();
    ObjectRef ref = ins.function(Counter::getPid).invoke();
    long pid = (long)YR.get(ref, 10000);

    ProcessBuilder pb = new ProcessBuilder("kill", "-9", String.valueOf(pid));
    pb.inheritIO();
    pb.start().waitFor();

    byte[] byteArray = YR.kv().get("myKey", 30);
    String recoverValue = new String(byteArray, StandardCharsets.UTF_8);
    Assert.assertEquals(recoverValue, "myValue");
    YR.Finalize();
}
```

样例 2：

```java

// 有状态函数中未定义 yrRecover
public class Counter {
    private int cnt = 0;

    public int add() {
        this.cnt += 1;
        return this.cnt;
    }

    public static long getPid () {
        String processName = ManagementFactory.getRuntimeMXBean().getName();
        long pid = Long.parseLong(processName.split("@")[0]);
        return pid;
    }

    public boolean saveState() throws YRException {
        YR.saveState();
        return true;
    }
}

public void main(String[] args) throws Exception {
    InstanceHandler ins = YR.instance(Counter::new).invoke();
    ObjectRef ref = ins.function(Counter::add).invoke();
    YR.get(ref, 10);

    ref = ins.function(Counter::saveState).invoke();
    YR.get(ref, 10);

    ref = ins.function(Counter::getPid).invoke();
    long pid = (long)YR.get(ref, 10000);

    ProcessBuilder pb = new ProcessBuilder("kill", "-9", String.valueOf(pid));
    pb.inheritIO();
    pb.start().waitFor();

    ref = ins.function(Counter::add).invoke();
    int res = (int)YR.get(ref, 10);
    Assert.assertEquals(res, 2);
    YR.Finalize();
}
```
