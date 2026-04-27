# Traefik Etcd Provider 配置说明

## 概述

AIO 镜像使用 Traefik 作为反向代理和 API 网关，通过 etcd provider 实现动态服务发现。

## 架构

```text
┌─────────────────┐      ┌──────────────┐      ┌─────────────────┐
│   Traefik       │──────│    etcd      │──────│  Services       │
│  (Priority 10)  │      │  (Port 32379)│      │  (动态注册)      │
└─────────────────┘      └──────────────┘      └─────────────────┘
```

## 启动顺序

由 Supervisor 管理，按以下顺序启动：

1. **Traefik** (priority=10) - 最先启动，如果 etcd 不可用会持续重试
2. **Runtime-Launcher** (priority=20) - gRPC 服务
3. **Yuanrong Master** (priority=30) - 启动 etcd 和其他服务

## etcd 配置

- **地址**: `127.0.0.1:32379`
- **Root Key**: `traefik`
- **TLS**: 禁用 (insecureSkipVerify: true)

## 服务注册格式

服务需要在 etcd 中注册以下格式的键：

```text
/traefik/http/services/{service-name}/loadBalancer/servers/{server-id}/url = "https://127.0.0.1:port"
/traefik/http/routers/{router-name}/rule = "PathPrefix(`/path`)"
/traefik/http/routers/{router-name}/service = "{service-name}"
```

## 静态路由

静态路由通过 `traefik/dynamic.yml` 配置，包括：

- `/terminal` - Terminal 服务
- `/api/*` - API 端点
- `/functions` - 函数管理
- `/healthz` - 健康检查

## 动态路由

动态路由通过 etcd 注册，例如：

```bash
# 注册一个新服务
etcdctl put /traefik/http/services/my-service/loadBalancer/servers/0/url "http://127.0.0.1:8080"
etcdctl put /traefik/http/routers/my-router/rule "PathPrefix(`/myservice`)"
etcdctl put /traefik/http/routers/my-router/service "my-service"
```

## 日志

- Traefik 日志: `/var/log/supervisor/traefik.log`
- Yuanrong 日志: `/var/log/supervisor/yuanrong.log`
- Supervisor 日志: `/var/log/supervisor/supervisord.log`
