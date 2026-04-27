# 分布式融合调用栈

在 openYuanrong 的分布式函数系统中，一次调用可能跨越多个实例、容器或物理节点。为帮助开发者快速定位异常源头，openYuanrong 提供了 **分布式融合调用栈（Distributed Stack）** 功能。

该功能将跨节点的 Java 调用路径透明地串联起来，并在发生异常时保留完整的异常类型与堆栈信息。特别地，当捕获到 `YRException` 异常时，用户可以通过`YRException.getCause()`方法获取到远程节点上抛出的**真实 Java 异常对象**（如 `NullPointerException`、`IllegalArgumentException` 等），无需解析日志或依赖外部监控。

## 背景与价值

在纯 Java 的分布式函数场景中，一个函数可能调用另一个部署在不同节点上的实例。若被调函数抛出异常，传统 RPC 框架通常仅返回通用错误码或模糊消息，导致调用方难以判断根本原因。

openYuanrong 的分布式融合调用栈解决了这一问题：

- **端到端 Java 异常透传**：远程 JVM 抛出的原始异常会被序列化并完整传递回调用方。
- **无缝集成 Java 异常机制**：调用方可直接使用 `getCause()`、`printStackTrace()` 等标准方法处理异常。
- **保留完整堆栈上下文**：包括类名、方法名、文件名和行号，便于精准定位问题代码。

## 限制与注意事项

- 仅支持 **Java 到 Java** 的函数调用。若调用非 Java 函数，`getCause()` 可能返回 `null` 或封装异常。
- 链式调用最高只支持10层，暂不支持更高层数。
- 被调函数抛出的异常必须实现 `java.io.Serializable`，否则无法透传。
- 自定义异常类需确保在调用方和被调方的 classpath 中均存在，且版本兼容。
- 自定义异常必须实现入参为String的构造函数。
- 敏感信息（如本地文件路径）在跨节点传输时会被自动脱敏。

## 使用方式

### 多层 Java 调用链示例

假设存在调用链：  
`Driver → InstanceA → InstanceB`

- 若 InstanceB` 抛出 `IOException`
- `InstanceA` 未捕获该异常
- `Driver` 捕获 `YRException` 后调用 `e.getCause()`，将直接获得 `IOException` 实例

openYuanrong 会自动将异常从 `InstanceB` 透传至 `Driver`，中间节点不改变异常类型。

以下代码展示了如何安全链式调用远程 Java 函数并获取真实异常：

```java
package com.example;

import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.call.InstanceHandler;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;

import java.io.IOException;
import java.nio.file.NoSuchFileException;
import java.util.List;
import java.nio.file.Files;
import java.nio.file.Paths;

public class Main {

    public static class MyYRApp {
        public static String failedSmallCall() throws YRException, IOException {
            String filePath = "not_exist.txt";
            // 真实的异常
            List<String> lines = Files.readAllLines(Paths.get(filePath));
            for (String line : lines) {
                System.out.println(line);
            }
            return "";
        }

        public static String remoteChainingCall() throws YRException, IOException {
            // 链式调用，异常会逐层透传
            InstanceHandler instance = YR.instance(Main.MyYRApp::new).invoke();
            ObjectRef ref1 = instance.function(MyYRApp::failedSmallCall).invoke();
            String res = (String) YR.get(ref1, 10000);
            return res + "-remoteChainingCall";
        }
    }

    public static void main(String[] args) throws YRException {
        YR.init(new Config());
        InstanceHandler instance = YR.instance(Main.MyYRApp::new).invoke();
        try {
            ObjectRef ref1 = instance.function(MyYRApp::remoteChainingCall).invoke();
            YR.get(ref1, 10000);
        }  catch (YRException e) {
            // 关键：通过 getCause() 获取远程抛出的真实 Java 异常
            Throwable cause = e.getCause();
            if (cause != null) {
                System.err.println("捕获到远程异常:");
                System.err.println("  类型: " + cause.getClass().getName());
                System.err.println("  消息: " + cause.getMessage());
            } else {
                System.err.println("调用失败，但未携带具体异常: " + e.getMessage());
            }
        } finally {
            YR.Finalize();
        }
    }
}
```

**远程异常捕获输出示例：**

```shell
捕获到远程异常:
  类型: java.nio.file.NoSuchFileException
  消息: null
java.nio.file.NoSuchFileException
        at com.example.MyYRApp.failedSmallCall(MyYRApp.java:11)
        at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
        at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
        at java.lang.reflect.Method.invoke(Method.java:498)
        at org.yuanrong.executor.FunctionHandler.invoke(FunctionHandler.java:361)
        at org.yuanrong.executor.FunctionHandler.execute(FunctionHandler.java:122)
        at org.yuanrong.codemanager.CodeExecutor.execute(CodeExecutor.java:81)
        at org.yuanrong.Entrypoint.runtimeEntrypoint(Entrypoint.java:102)
        at org.yuanrong.runtime.server.RuntimeServer.main(RuntimeServer.java:51)
        at com.example.MyYRApp.remoteChainingCall(MyYRApp.java:23)
```

## 技术实现说明

- **异常序列化**：所有可序列化的 Java 异常（实现 `Serializable`）均可被透传。
- **堆栈增强**：远程异常的堆栈信息会被完整保留，并附加调用路径元数据（如目标节点 ID）。
- **类型安全**：反序列化后的异常对象保持原始类类型，支持 `instanceof` 判断。

例如：

```java
if (e.getCause() instanceof IllegalArgumentException) {
    // 可安全进行类型特异性处理
}
```

## 相关 API

| 类/方法                              | 说明                                                         |
|  |  |
| `org.yuanrong.exception.YRException` | 远程调用失败时抛出的封装异常                                 |
| `YRException.getCause()`             | 获取被封装的真实 Java 异常（可能为 `NullPointerException`、`IOException` 等） |

通过分布式融合调用栈，openYuanrong 让 Java 开发者在分布式环境中依然能享受与本地开发一致的异常调试体验——**“所见即所得，异常即根源”**。
