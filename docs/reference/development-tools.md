# 常用工具安装

本节介绍如何在 Linux 环境中安装一些常用的开发工具。

## MinIO Client(mc)

[Minio Client](https://minio.org.cn/docs/minio/linux/reference/minio-mc.html){target="_blank"} 简称 mc, 是 minio 服务器的客户端。

### 安装

[下载](https://dl.min.io/client/mc/release){target="_blank"}安装包中的 `mc` 文件到主机，执行以下命令安装：

```bash
chmod +x mc
sudo mv ./mc /usr/local/bin/mc
```

### 使用

初始化配置：

```bash
查看 openYuanrong 集群中的 MinIO 服务地址
echo http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc minio -o jsonpath='{.spec.ports[0].nodePort}')
```

```bash
# mys3: 服务器别名
# http://x.x.x:30110: 服务器地址
# <AK> <SK>: 配置的 MinIO AK/SK
mc alias set mys3 http://x.x.x:30110 <AK> <SK>
# 查看 mys3 信息
mc admin info mys3
```

常用命令

```bash
# 查看 bucket 列表，其中 mys3 是服务端的别名
mc ls mys3

# 查看某个 bucket 下的文件
mc ls mys3/this-bucket

# 创建 bucket
mc mb mys3/this-bucket

# 删除 bucket
mc rm mys3/this-bucket

# 上传文件
mc cp ./demo.zip mys3/this-bucket/demo.zip

# 下载文件
mc cp mys3/this-bucket/demo.zip ./demo.zip

# 删除文件
mc rm mys3/this-bucket/demo.zip

```
