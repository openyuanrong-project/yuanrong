# AI Agent 会话与亲和性调度

AI Agent 会话功能专为交互式应用场景（如 AI 智能体、多轮对话）设计。它支持函数执行过程中的主动等待与外部唤醒，并确保同一会话内的多次请求能够路由到同一个执行实例，从而实现低延迟的交互体验。

## 会话调度机制

AI Agent 会话的调度逻辑由 `faasscheduler` 模块负责，具有以下核心特点：

### 1. 会话亲和性 (Session Affinity)

当一个请求携带 `sessionId` 时，调度系统会尝试将其路由到该会话已经绑定的实例上。

- **首次请求**: 调度系统选择一个合适的实例，并建立 `sessionId` 与 `instanceId` 的绑定关系。
- **后续请求**: 只要绑定关系有效，所有相同 `sessionId` 的请求都会定向到同一实例。

### 2. 弱亲和性与持久化

AI Agent 会话默认为弱亲和（`TTL` 为 ``0``）。

- **生命周期**: 当会话下的所有租约都释放后，绑定关系可能会失效。
- **状态恢复**: 会话的上下文数据（如历史记录）持久化在分布式数据系统中。当新请求到达且原实例不可用时，新实例会从数据系统重新加载会话状态。

## SDK 使用说明

在启用了 AI Agent 会话的函数中，可以通过 `Context` 获取 `SessionService` 来操作会话对象。

### Java SDK

#### 核心接口：Session

`Session` 对象提供了会话内同步和状态管理的能力。

- **`wait(long timeoutMs)`**: 挂起当前执行线程，等待输入。
- **`notify(JsonObject payload)`**: 唤醒正在 `wait` 的线程。
- **`getInterrupted()`**: 检查当前会话是否已被外部中断。

#### Java 使用示例

```java
import com.google.gson.JsonObject;

public Object handle(Context ctx, JsonObject input) {
    Session sess = ctx.getSessionService().loadSession(ctx.getSessionId());

    // 检查是否为 notify 请求，假设用户在 payload 字段中传入通知内容
    if (input.has("action") && "notify".equals(input.get("action").getAsString())) {
        sess.notify(input.getAsJsonObject("payload"));
        return "Notified";
    }

    JsonObject userInput = sess.wait(60000); 
    
    if (userInput == null) return "Timeout";
    if (sess.getInterrupted()) return "Interrupted";
    
    return "Got: " + userInput.toString();
}
```

### Python SDK

在 Python 函数实例中，`yr` 模块提供了类似的会话操作能力。

#### 核心接口：`yr.SessionService`

- **`session.wait_for_notify(timeout_ms)`**: 阻塞当前协程/线程，等待通知。
- **`session.notify(payload)`**: 发送通知。
- **`session.is_interrupted()`**: 检查会话是否被中断。

#### Python 使用示例

```python
import yr

def handle(ctx, input):
    session_id = ctx.get_session_id()
    session = ctx.get_session_service().load_session(session_id)

    if input.get("action") == "notify":
        session.notify(input.get("payload"))
        return "Notified"

    # 等待通知
    user_input = session.wait_for_notify(60000)
    
    if user_input is None:
        return "Wait timeout"
        
    if session.is_interrupted():
        return "Session Interrupted"
        
    return f"Received: {user_input}"
```

## 完整用例：多轮对话 AI Agent

以下是一个完整的 Java 示例，演示了如何利用 `wait`/`notify` 实现一个简单的多轮交互。

```java
import org.yuanrong.services.Context;
import org.yuanrong.services.session.Session;
import org.yuanrong.services.session.SessionService;
import com.google.gson.JsonObject;
import java.util.ArrayList;
import java.util.List;

public class SimpleAgent {
    public Object handle(Context ctx, JsonObject input) {
        SessionService sessService = ctx.getSessionService();
        Session sess = sessService.loadSession(ctx.getSessionId());

        // 1. 处理通知/唤醒请求
        if (input.has("action") && "notify".equals(input.get("action").getAsString())) {
            // 提取真正的通知载荷进行唤醒
            sess.notify(input.getAsJsonObject("payload"));
            return null; // Notify 请求通常不需要返回业务结果
        }

        // 2. 主执行流程
        try {
            ctx.getLogger().log("Agent Started, SessionID: " + ctx.getSessionId());
            
            // 第一轮交互
            ctx.getStream().write("你好！我是 AI 助手，请问有什么可以帮您？\n");
            JsonObject input1 = sess.wait(30000); // 等待用户输入 30 秒
            if (input1 == null) return "等待超时";
            
            String msg1 = input1.get("message").getAsString();
            ctx.getStream().write("收到指令：" + msg1 + "\n正在为您处理...\n");
            
            // 模拟处理过程并更新历史
            List<String> history = new ArrayList<>(sess.getHistories());
            history.add("User: " + msg1);
            sess.setHistories(history);
            
            // 第二轮交互
            ctx.getStream().write("处理完成。您还有其他问题吗？\n");
            JsonObject input2 = sess.wait(30000);
            if (input2 == null) return "等待超时";

            if (sess.getInterrupted()) {
                return "会话已被外部中断";
            }

            String msg2 = input2.get("message").getAsString();
            ctx.getStream().write("好的，已收到您的进一步要求：" + msg2 + "\n");

        } catch (Exception e) {
            ctx.getLogger().log("Error: " + e.getMessage());
        }

        return "Session Completed";
    }
}
```

### 交互流程示意图

第一轮交互：逻辑挂起

1. **客户端** 发起 `POST /invocations` (携带 `SessionID`: ``S1``)
2. **函数实例** 接收请求，执行业务逻辑，直到调用 `sess.wait()`
3. **函数实例** 释放会话锁，执行线程进入挂起状态，等待唤醒

第二轮交互：唤醒与继续

1. **客户端** 发起第二个请求 `POST /invocations` (相同 `SessionID`: ``S1``, `Action`: ``notify``)
2. **函数实例** 接收请求，由于会话亲和性，该请求进入同一实例
3. **函数实例** 执行 `sess.notify(payload)`，唤醒之前挂起的线程
4. **函数实例** 通知请求处理完成并返回 ``200 OK``
5. **函数实例** (原线程) 获取到 `notify` 传递的数据，继续执行后续逻辑
6. **函数实例** (原线程) 完成处理，向客户端返回第一轮请求的最终结果
