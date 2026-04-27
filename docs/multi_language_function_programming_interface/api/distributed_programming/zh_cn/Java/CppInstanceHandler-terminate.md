# CppInstanceHandler terminate

包名：`org.yuanrong.call`。

## 接口说明

### public void terminate() throws YRException

[CppInstanceHandler](CppInstanceHandler.md) 类的成员方法，用于回收云上 cpp 函数实例。

:::{Note}

当前 kill 请求的默认超时时间为 30 s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试。

:::

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
cppInstanceHandler.terminate();
```

- 抛出：

   - **YRException** - 回收实例失败抛出该异常，具体错误详见错误信息描述。

### public void terminate(boolean isSync) throws YRException

[CppInstanceHandler](CppInstanceHandler.md) CppInstanceHandler 类的成员方法，用于回收云上 cpp 函数实例。

支持同步或异步 terminate。

:::{Note}

在不开启同步 terminate 时，当前 kill 请求的默认超时时间为 30 s，在磁盘高负载、etcd 故障等场景下，kill 请求处理时间可能超过 30 s，会导致接口抛出超时的异常，由于 kill 请求存在重试机制，用户可以选择在捕获超时异常后不处理或者进行重试；在开启同步 terminate 时，该接口会阻塞等待，直至实例完全退出。

:::

- 参数：

   - **isSync** - 是否开启同步。若为 true，表示向 function-proxy 发送信号量为  killInstanceSync 的 kill 请求，内核同步 kill 实例；若为 false，表示向 function-proxy 发送信号量为 killInstance 的 kill 请求，内核异步 kill 实例。

- 抛出：

   - **YRException** - 回收实例失败抛出该异常，具体错误详见错误信息描述。
