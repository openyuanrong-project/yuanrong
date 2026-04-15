# 基于 AI Agent 会话实现多轮对话助手

本案例展示如何利用 openYuanrong 的 AI Agent 会话（AI Agent Session）能力，构建一个支持多轮交互的智能对话助手。通过 `wait` 和 `notify` 机制，函数可以在执行过程中挂起并等待用户输入，从而实现复杂的交互逻辑。

## 场景描述

在传统的 FaaS 模型中，函数通常是触发式执行且“运行即结束”的。对于多轮对话场景，开发者通常需要自行管理复杂的上下文状态（如 Redis 或数据库）。

openYuanrong 的 AI Agent 会话功能提供了以下优势：

- **自动状态管理**：会话上下文随会话生命周期自动持久化和加载。
- **执行流挂起**：支持在函数内部通过 `wait` 挂起，等待同一会话的后续请求唤醒。
- **低延迟响应**：通过会话亲和性调度，确保同一会话的请求路由到同一实例，减少冷启动和状态同步开销。

## 开发步骤

### 1. 开启函数会话功能

在注册函数时，设置 `enable_agent_session` 为 ``true``。

```bash
# 使用 yr 命令行工具或 REST API 注册
curl -X POST http://{meta_service}/serverless/v1/functions \
-H "Content-Type: application/json" \
-d '{
    "name": "0@ai@dialogue-agent",
    "runtime": "java8",
    "handler": "org.example.DialogueAgent.handle",
    "enable_agent_session": true,
    "cpu": 1000,
    "memory": 1024
}'
```

### 2. 编写函数逻辑 (Java)

使用 Java SDK 实现一个简单的多轮 Q&A 逻辑。

```java
package org.example;

import org.yuanrong.services.Context;
import org.yuanrong.services.session.Session;
import org.yuanrong.services.session.SessionService;
import com.google.gson.JsonObject;
import java.util.ArrayList;
import java.util.List;

public class DialogueAgent {
    public Object handle(Context ctx, JsonObject input) {
        SessionService sessService = ctx.getSessionService();
        Session sess = sessService.loadSession(ctx.getSessionId());

        // 1. 如果是通知请求（由后续调用触发），则执行唤醒逻辑
        if (input.has("action") && "notify".equals(input.get("action").getAsString())) {
            sess.notify(input.getAsJsonObject("payload"));
            return null; // Notify 请求不返回结果给当前触发者
        }

        // 2. 主会话执行流程
        try {
            // 第一轮
            ctx.getStream().write("你好！我是您的 AI 助手。请问您怎么称呼？\n");
            JsonObject nameInput = sess.wait(60000); // 挂起并等待输入，超时 60s
            if (nameInput == null) return "等待称呼超时";
            
            String userName = nameInput.get("message").getAsString();
            
            // 更新会话历史
            List<String> history = new ArrayList<>(sess.getHistories());
            history.add("User Name: " + userName);
            sess.setHistories(history);

            // 第二轮
            ctx.getStream().write(userName + "，很高兴见到你！请问有什么我可以帮您的吗？\n");
            JsonObject taskInput = sess.wait(60000);
            if (taskInput == null) return "等待任务指令超时";

            String task = taskInput.get("message").getAsString();
            ctx.getStream().write("好的，正在为您处理任务：" + task + "...\n");
            
            // 模拟任务处理...
            Thread.sleep(2000);
            
            ctx.getStream().write("处理完成！再见，" + userName + "。");

        } catch (Exception e) {
            ctx.getLogger().log("Error: " + e.getMessage());
        }

        return "Session Finished";
    }
}
```

## 客户端交互流程

步骤 A: 启动会话并开始第一轮

1. **客户端**: 调用 `POST /invocations` (SessionID: S1)
2. **网关**: 转发请求到绑定的实例
3. **函数实例**: 执行业务逻辑并进入 `sess.wait()`，执行线程挂起

步骤 B: 发送通知数据进行唤醒

1. **客户端**: 调用 `POST /invocations` (相同 SessionID, Action: notify)
2. **网关**: 将请求分发到同一个实例（会话亲和性）
3. **函数实例**: 执行 `sess.notify(payload)`，唤醒之前挂起的线程
4. **函数实例**: 向网关返回 Notify 操作成功响应

唤醒与最终响应

1. **函数实例**: 原线程被成功唤醒，获取到数据后继续处理后续逻辑
2. **函数实例**: 完成全部处理逻辑，向客户端返回第一轮请求的最终结果

## 总结

通过 AI Agent 会话，开发者可以像编写单机同步程序一样编写多轮交互逻辑，极大简化了分布式上下文管理的复杂度。结合 openYuanrong 的会话亲和调度，系统能够高效处理通知请求，确保交互的实时性。
