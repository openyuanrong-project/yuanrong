# InstanceHandler terminate

包名：`org.yuanrong.call`。

## 接口说明

### public void terminate() throws YRException

[InstanceHandler](InstanceHandler.md) 类的成员方法，用于回收云上 Java 函数实例。 当前 kill 请求的默认超时时间为 30 s，在磁盘高负载 etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。

:::{Note}

当前终止请求的默认超时时间为 30 秒。在高磁盘负载和 etcd 故障等场景下，终止请求的处理时间可能会超过 30 秒，从而导致接口抛出超时异常。由于终止请求具有重试机制，用户可以选择在捕获超时异常后不进行处理，或者进行重试。

:::

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate();
```

- 抛出：

   - **YRException** - 统一抛出的异常类型。

### public void terminate(boolean isSync) throws YRException

[InstanceHandler](InstanceHandler.md) 类的成员方法，用于回收云上 Java 函数实例。支持同步或异步 terminate。在不开启同步 terminate 时，当前 kill 请求的默认超时时间为 30 s，在磁盘高负载 etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。在开启同步 terminate 时，该接口会阻塞等待，直至实例完全退出。

#### 注意

当未启用同步终止时，当前终止请求的默认超时时间为 30 秒。在高磁盘负载或 etcd 故障等场景下，终止请求的处理时间可能会超过 30 秒，从而导致接口抛出超时异常。由于终止请求具有重试机制，用户可以选择在捕获超时异常后不进行处理，或者进行重试。
当启用同步终止时，接口会阻塞，直到实例完全退出。

```java

InstanceHandler InstanceHandler = YR.instance(Counter::new).invoke(1);
InstanceHandler.terminate(true);
```

- 参数：

   - **isSync** - 是否开启同步。若为 true，表示向 function-proxy 发送信号量为 killInstanceSync 的 kill 请求， 内核同步 kill 实例；若为 false，表示向 function-proxy 发送信号量为 killInstance 的 kill 请求，内核异步 kill 实例。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
