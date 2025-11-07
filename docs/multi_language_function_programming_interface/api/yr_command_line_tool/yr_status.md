# `status`

查询主机部署的 openYuanrong 集群状态。

## 用法

```shell
yr status [flags]
```

## 参数

* `-h`：输出命令帮助信息。
* `--etcd_endpoint`：ETCD 服务地址。使用`yr start`命令输出的`Cluster master info`信息，拼接 `{etcd_ip}:{etcd_port}`字段为"IP:Port"的形式。
* `--etcd_tls_enable`：ETCD 服务启用 TLS。若`yr start`命令参数`etcd_auth_type`不为`Noauth`，需要使用该参数。该参数无需传入值。
* `--etcd_ca_file`：ETCD 服务 CA 文件路径。启用`etcd_tls_enable`时该参数必选，参考 [部署参数表](../../../deploy/deploy_processes/parameters.md)
* `--etcd_client_key_file`：ETCD 客户端密钥文件路径。启用`etcd_tls_enable`时该参数必选，参考 [部署参数表](../../../deploy/deploy_processes/parameters.md)
* `--etcd_client_cert_file`：ETCD 客户端证书文件路径。启用`etcd_tls_enable`时该参数必选，参考 [部署参数表](../../../deploy/deploy_processes/parameters.md)

## Example

```shell
[yuanrong@example ~]$ yr status --etcd_endpoint xxx.xxx.xxx.xxx:xxxxxx
YuanRong is running normally,
YuanRong cluster addresses:
              functionsystem: x.x.x.x:30839
                  datasystem: x.x.x.x:39133

YuanRong cluster status:
  current running agents: 1
```
