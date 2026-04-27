# yrShutdown

:::{Note}

Constraints:

- When users define their own `yrShutdown`, the function signature must match the interface prototype. Otherwise, the cloud will not be able to find the corresponding function to execute graceful shutdown.

- If the execution time of `yrShutdown` exceeds `gracePeriodSeconds` seconds, the cloud instance will not wait for the function to complete and will be immediately reclaimed.

:::

## Interface description

### public void yrShutdown(int gracePeriodSeconds)

Users can use this interface to perform data cleanup or data persistence operations.

In actors executed in the cloud, add a graceful shutdown function named `yrShutdown`. The function supports calling the openYuanrong interface. The `yrShutdown` function will be executed in the following scenarios:

1. When the runtime receives a shutdown request (for example, when a user on-premises calls the terminate interface).

2. When the runtime captures a Sigterm signal (for example, when the cluster executes a scaling-down policy and the runtime-manager captures the Sigterm signal and forwards it to the runtime).

```java

// Define `yrShutdown` in the stateful function.
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
    // Use Case 1: GRACEFUL_SHUTDOWN_TIME is not configured, using the system-configured graceful shutdown time.
    YR.kv().del("myKey");
    InstanceHandler ins = YR.instance(Counter::new).invoke();
    ObjectRef ref = ins.function(Counter::add).invoke();
    YR.get(ref, 10);
    ins.terminate();
    // On-premises, using the `kv` interface to retrieve the value mapped by `my_key` will return `'my_value'`.
    byte[] byteArray = YR.kv().get("myKey", 30);
    String shutdownValue = new String(byteArray, StandardCharsets.UTF_8);
    Assert.assertEquals(shutdownValue, "myValue");

    // Use Case 2: Configure GRACEFUL_SHUTDOWN_TIME.
    YR.kv().del("myKey");
    InvokeOptions option = new InvokeOptions();
    option.getCustomExtensions().put("GRACEFUL_SHUTDOWN_TIME","10");
    InstanceHandler ins2 = YR.instance(Counter::new).options(option).invoke();
    ObjectRef ref2 = ins2.function(Counter::add).invoke();
    YR.get(ref2, 10);
    ins2.terminate();
    // On-premises, using the kv interface to retrieve the value mapped by my_key will return 'my_value'.
    byte[] byteArray2 = YR.kv().get("myKey", 30);
    String shutdownValue2 = new String(byteArray2, StandardCharsets.UTF_8);
    Assert.assertEquals(shutdownValue2, "myValue");

    YR.Finalize();
}
```

- Parameters:

   - **gracePeriodSecond** (int) – Function execution timeout, in seconds. Default value: ``60``.
