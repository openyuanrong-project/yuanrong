# 元戎沙箱能力使用指南 - OpenClaw & Claude Code

## 一、环境信息

### 1.1 元戎沙箱试用环境
```
https://1.95.144.130:8888/terminal
```

### 1.2 试用Token

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzIxNzAyNTcsInJvbGUiOiJkZXZlbG9wZXIiLCJzdWIiOiJkZWZhdWx0In0.YmNmM2M3NDYxZTUyMjdmYmMwMTZlMTEzM2RjZTc1YTJlMDUzYjNiZmRmNjUyMjgxNTEyYmE0MmEyODkyZDcyMg
```
---
## 二、Claude Code 使用方法

### 2.1 环境变量配置

```bash
# 设置 智谱 认证 Token
export ANTHROPIC_AUTH_TOKEN=1ff46be15f4d447ba5c03a3e97b06fc5.kAkTe5mHTsQIAFa2

# 设置 智谱 API 基础地址
export ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic
```

### 2.2 启动 Claude Code

```bash
# 直接启动
claude

# 或带参数启动
claude [项目目录]
```

---

## 三、OpenClaw 使用方法

### 3.1 环境变量配置

```bash
# 设置 智谱 API Key
export ZAI_API_KEY=1ff46be15f4d447ba5c03a3e97b06fc5.kAkTe5mHTsQIAFa2
```

### 3.2 初始化配置

```bash
# 非交互式初始化，自动接受风险并选择 zai-cn 认证
openclaw onboard \
    --non-interactive \
    --accept-risk \
    --auth-choice zai-cn \
    --skip-channels \
    --skip-skills \
    --skip-ui
```

### 3.3 启动 Gateway 服务

```bash
# 后台启动 gateway 服务，并将日志输出到文件
openclaw gateway >> openclaw.log 2>&1 &
```

### 3.4 启动 TUI 界面

```bash
# 启动 OpenClaw 终端用户界面
openclaw tui
```

---

## 四、本地部署快速启动

### 4.1 启动容器

```bash
docker run -d \
  -p 8888:8888 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  swr.cn-southwest-2.myhuaweicloud.com/openyuanrong/openyuanrongaio:latest
```

启动后访问 Web 终端：`https://127.0.0.1:8888/terminal`

### 4.2 生成访问 Token

**方式一：使用私钥生成（需知道私钥）**

```bash
# 生成 Token 私钥（LITEBUS_DATA_KEY 的明文值）
python3 -c "print('openyuanrong.org'.encode().hex().upper())"
```

**方式二：在容器内直接获取 Token**

```bash
# 进入容器后执行，iam-address 替换为实际 IAM 服务地址
yrcli token-require --tenant-id default --role developer --iam-address 172.17.0.3:31112
```

### 4.3 容器内使用 Claude Code / OpenClaw

进入 Web 终端后，参照第二节和第三节配置环境变量并启动即可：

```bash
# Claude Code
export ANTHROPIC_AUTH_TOKEN=<your-token>
export ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic
claude

# OpenClaw
export ZAI_API_KEY=<your-api-key>
openclaw onboard --non-interactive --accept-risk --auth-choice zai-cn \
    --skip-channels --skip-skills --skip-ui
openclaw gateway >> openclaw.log 2>&1 &
openclaw tui
```