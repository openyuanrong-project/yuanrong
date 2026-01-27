# API

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   distributed_programming/zh_cn/index
   function_service/index
   yr_command_line_tool/index
   error_code
```

API 分类如下：

- [单机程序分布式并行化 API](./distributed_programming/zh_cn/index.md)：提供开发无状态函数、有状态函数的接口。
- [函数服务 API](./function_service/index.md)：提供函数服务的开发、调用及管理接口。

## 如何调用 REST API

调用 REST API 需要构造请求 URI，它通常由请求协议、服务端点及资源路径三部分组成。默认部署的 openYuanrong 集群采用 http 请求协议，服务端点由 frontend 和 meta service 组件提供。以注册函数 REST API 为例，完整的 URI 为：`http://{meta service endpoint}/serverless/v1/functions`。

(api-frontend-endpoint)=

### 获取 frontend 组件服务端点

frontend 组件提供了调用服务、订阅流服务等 REST API，服务端点可通过以下方式获取。

- 主机上部署 openYuanrong 时，主节点成功部署将打印如下内容，组合 master_ip 和 frontend_port，即获得 frontend 组件的服务端点：192.168.0.1:30556。部署后，您也可以通过主节点启动信息文件获取，默认路径为 `/tmp/yr_sessions/latest/master.info`。

   ```text
   Cluster master info:
      local_ip:192.168.0.1,master_ip:192.168.0.1,etcd_ip:192.168.0.1,etcd_port:10012,global_scheduler_port:17979,ds_master_port:12577,cluster_deployer_port:22775,etcd_peer_port:15606,bus-proxy:37545,bus:34577,ds-worker:38664,metaservice_port:37774,frontend_port:30556,
   ```

- K8s 上部署 openYuanrong 时，在 K8s 主节点通过 kubectl 命令获取。

   ```bash
   echo "$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc faas-frontend-lb -o jsonpath='{.spec.ports[0].nodePort}')"
   ```

(api-meta-service-endpoint)=

### 获取 meta service 组件服务端点

meta service 组件提供了函数管理、资源池管理等 REST API，服务端点可通过以下方式获取。

- 主机上部署 openYuanrong 时，主节点成功部署将打印如下内容，组合 master_ip 和 metaservice_port，即获得 meta service 组件的服务端点：192.168.0.1:37774。部署后，您也可以通过主节点启动信息文件获取，默认路径为 `/tmp/yr_sessions/latest/master.info`。

   ```text
   Cluster master info:
      local_ip:192.168.0.1,master_ip:192.168.0.1,etcd_ip:192.168.0.1,etcd_port:10012,global_scheduler_port:17979,ds_master_port:12577,cluster_deployer_port:22775,etcd_peer_port:15606,bus-proxy:37545,bus:34577,ds-worker:38664,metaservice_port:37774,frontend_port:30556,
   ```

- K8s 上部署 openYuanrong 时，在 K8s 主节点通过 kubectl 命令获取。

   ```bash
   echo "$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc meta-service -o jsonpath='{.spec.ports[0].nodePort}')"
   ```
