# 流式响应

在大语言模型推理场景中，流式返回能够在模型生成文本的同时逐步将结果发送给客户端，从而实现实时的交互体验。

函数服务支持响应流式数据，openYuanrong 提供了两种流式响应 REST API：[订阅流服务](../../api/function_service/stream_invocation.md)和[调用服务](../../api/function_service/function_invocation.md)时流式返回。

## 订阅流服务

使用单机程序分布式并行化接口中的数据流 API 生产的流，可以通过调用[订阅流服务](../../api/function_service/stream_invocation.md) REST API 消费。使用该接口需要注意以下几点：

- 流生产者和订阅者保持一样的流名称。
- 业务逻辑上确保先订阅流再生产数据，避免生产的数据无法被全部消费。
- 建议流生产者和订阅者协商特定的结束符表达流数据传递结束，便于订阅者消费完数据后主动关闭。

通过[在函数服务中使用流](../../examples/use_stream.md)查看完整使用示例。

## 调用服务时流式返回

调用 openYuanrong 服务时流式返回基于 SSE(Server-Send Events) 协议实现。在函数服务中通过上下文接口中的 `context.getStream()` 关联流，发送流数据。一个流式返回的 Java 函数服务示例如下。

```java
package org.yuanrong.demo;

import org.yuanrong.services.runtime.Context;
import com.google.gson.JsonObject;

public class Demo {
    public String handler(JsonObject jsonObject, Context context) throws Exception {
        try {
            Stream stream = context.getStream();
            JsonObject obj = new JsonObject();
            obj.addProperty("name","handler");
            obj.addProperty("event","this is a stream");
            stream.write(obj);
            stream.write(new JsonObject());
        } catch (Exception e) {
            System.out.println("executor exception occurred");
        }
        return "ok";
    }
}
```
