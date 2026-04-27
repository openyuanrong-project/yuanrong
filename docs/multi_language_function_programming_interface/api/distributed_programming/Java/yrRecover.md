# yrRecover

:::{Note}

Constraint: When users define their own `yrRecover`, the function signature must match the interface prototype. Otherwise, the cloud will not be able to find the corresponding function and execute it.

:::

## Interface description

### public void yrRecover()

Users can use this interface to perform state recovery, initialization of state, and other operations.
In actors executed in the cloud, add a hook function named `yrRecover`. The function supports calling the openYuanrong interface. The `yrRecover` function will be executed in the following scenarios:

1. When the runtime receives a recover request (for example, when a user on-premises configures `recoverRetryTimes > 0` in `InvokeOptions`, and the cloud instance exits abnormally).

Case 1:

```java

// Define yrRecover in the stateful function.
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

Case 2:

```java

// yrRecover is not defined in the stateful function.
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
