# 安全通信

openYuanrong 支持 TLS 1.2、TLS 1.3 协议用于外部及组件之间的安全通信，本节将介绍如何生成相关证书。

所需证书文件及目录规划如下：

```text
${Your_Workspace}
├── etcd // etcd 证书目录
│   ├── ca.crt
│   ├── client.crt
│   ├── client.key
│   ├── server.crt
│   └── server.key
├── yr // openYuanrong 函数系统组件证书目录
│   ├── ca.crt
│   ├── module.crt
│   └── module.key
└── curve // openYuanrong 数据系统 curve 证书目录
    ├── worker.key
    ├── worker.key_secret
    ├── client.key
    ├── client.key_secret
    └── worker_authorized_clients
        ├── client.key
        └── worker.key
```

## 环境准备

一台 Linux 主机。

### 安装 openssl

我们使用 openssl 生成 etcd 及 openYuanrong 函数系统相关证书。

```shell
yum install openssl -y

# 检查安装版本。
openssl version
```

### 安装 ZeroMQ

openYuanrong 数据系统组件使用 ZeroMQ 作为 RPC 通信框架，公私钥依赖 ZeroMQ 所提供的接口生成，同时 ZeroMQ 的安全功能又依赖 libsodium。

[下载](https://github.com/jedisct1/libsodium/releases/tag/1.0.18-RELEASE){target="_blank"} 1.0.18 版本包 `libsodium-1.0.18.tar.gz`，参考如下命令安装 libsodium。

```shell
tar -zxvf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make && make install
```

[下载](https://github.com/zeromq/libzmq/releases/tag/v4.3.5){target="_blank"} 4.3.5 版本包 `zeromq-4.3.5.tar.gz`，参考如下命令安装 ZeroMQ。

```shell
tar -zxvf zeromq-4.3.5.tar.gz
cd zeromq-4.3.5
mkdir build
cd build
cmake .. -DENABLE_CURVE:BOOL=ON -DWITH_LIBSODIUM:BOOL=ON -DWITH_LIBSODIUM_STATIC:BOOL=ON
make && make install
```

### 创建工作目录

创建一个工作目录用于保存生成的证书文件，例如 `/opt/security_deployment`，后续操作将以 `${WorkSpace}` 表示您的工作目录。

```shell
mkdir -p /opt/security_deployment
export WorkSpace=/opt/security_deployment
```

## 生成 etcd 证书等文件

以下步骤将生成 etcd 的 CA 证书、服务端证书、服务端私钥、客户端证书和客户端私钥共 5 个文件。

### 生成 etcd CA 证书

执行如下命令，生成 CA 证书文件 `ca.crt`。

```shell
mkdir -p ${WorkSpace}/etcd
cd ${WorkSpace}/etcd
openssl genrsa -out ca.key 2048

# 证书CN 可根据实际配置修改，默认 etcd-ca。days 为证书有效期（天数），证书过期后需要重新生成。
openssl req -x509 -new -nodes -key ca.key -subj "/CN=etcd-ca" -days 10000 -out ca.crt
```

### 生成 etcd 服务端私钥及证书

创建 `etcd.conf` 文件, 其中 alt_names 配置项的值 `127.0.0.1` 替换为您 openYuanrong 集群主节点 IP。

```shell
cd ${WorkSpace}/etcd
cat >etcd.conf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no
[req_distinguished_name]
CN = etcd-ca
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
[alt_names]
IP.1 = 127.0.0.1
DNS.1 = 127.0.0.1
EOF
```

依次执行命令生成服务端私钥 `server.key` 以及服务端证书 `server.crt` 文件。

```shell
cd ${WorkSpace}/etcd
openssl genrsa -out server.key 2048

# 证书CN 可根据实际配置修改，这里默认 etcd-server。
openssl req -new -key server.key -subj "/CN=etcd-server" -out server.csr -config etcd.conf

# days 为证书有效期（天数），证书过期后需要重新生成。
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -extfile etcd.conf -extensions v3_req
```

### 生成 etcd 客户端私钥和证书

依次执行命令生成客户端私钥 `client.key` 以及客户端证书 `client.crt` 文件。

```shell
cd ${WorkSpace}/etcd
openssl genrsa -out client.key 2048

# 证书CN 可根据实际配置修改，这里默认 etcd-client。
openssl req -new -key client.key -subj "/CN=etcd-client" -out client.csr

# days 为证书有效期（天数），证书过期后需要重新生成。
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 10000
```

### 保存 etcd 证书文件

至此已完成 etcd 所需证书文件的生成，执行如下命令，保存这些文件到我们规划好的目录。

```shell
mkdir -p ${WorkSpace}/cert/etcd
cd ${WorkSpace}/etcd
cp -ar ca.crt client.crt client.key  server.key server.crt ${WorkSpace}/cert/etcd
```

## 生成 openYuanrong 函数系统组件证书等文件

以下步骤将生成 openYuanrong 函数系统组件的 CA 证书、私钥和证书共 3 个文件。

### 创建配置文件

执行以下命令创建配置文件, 其中 alt_names 配置项的值 `127.0.0.1` 替换为您 openYuanrong 集群主节点 IP。

```shell
mkdir -p ${WorkSpace}/yr
cd ${WorkSpace}/yr
rm -rf demoCA 2>/dev/null
mkdir -p ./demoCA/newcerts/
touch ./demoCA/index.txt
touch ./demoCA/index.txt.attr
echo 01 > ./demoCA/serial
if [ ! -f  ./demoCA/v3.ext ];
then
cat >./demoCA/v3.ext << EOF
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[ v3_ca ]
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always,issuer
basicConstraints = critical,CA:true
[ alt_names ]
DNS.1 = test
DNS.2 = 127.0.0.1
IP.1 = 127.0.0.1
EOF
fi
```

### 生成 openYuanrong 函数系统组件 CA 证书

执行如下命令，生成 CA 证书文件 `ca.crt`。

```shell
cd ${WorkSpace}/yr
openssl genrsa -out ca.key 2048

# subj 可根据实际配置修改。
openssl req -new -key ca.key -out ca.csr -subj "/C=CN/ST=zhejiang/L=hangzhou/O=ha/OU=Personal/CN=rootCA"

# days 为证书有效期（天数），证书过期后需要重新生成。
openssl x509 -req -days 10000 -extfile ./demoCA/v3.ext -extensions v3_ca -signkey ca.key -in ca.csr -out ca.crt
```

### 生成 openYuanrong 函数系统组件私钥和证书

依次执行命令生成私钥 `module.key` 以及证书 `module.crt` 文件。

```shell
cd ${WorkSpace}/yr
openssl genrsa -out module.key 2048

# subj 可根据实际配置修改
openssl req -new -key module.key -out module.csr -subj "/C=CN/ST=zhejiang/L=hangzhou/O=ha/OU=Personal/CN=test"
openssl ca -extfile ./demoCA/v3.ext -days 10000 -extensions v3_req -in module.csr -out module.crt -cert ca.crt -keyfile ca.key -notext -batch
```

### 保存 openYuanrong 函数系统组件证书文件

至此已完成 openYuanrong 函数系统组件所需证书文件的生成，执行如下命令，保存这些文件到我们规划好的目录。

```shell
mkdir -p ${WorkSpace}/cert/yr
cd ${WorkSpace}/yr
cp -ar ca.crt module.crt module.key ${WorkSpace}/cert/yr
```

## 生成 openYuanrong 数据系统组件 Curve 证书文件

### 生成 worker 及 client 公私钥

我们将使用 ZeroMQ 的接口 `zmq_curve_keypair` 生成公私钥，详细接口介绍参考 [Zmq 接口文档](http://api.zeromq.org/4-0:zmq-curve-keypair){target="_blank"}。

创建工作目录。

```shell
mkdir -p ${WorkSpace}/yr_data_system
cd ${WorkSpace}/yr_data_system
```

在当前目录下新建 `create_cert.cpp` 文件，内容如下。

```c++
extern "C" {
#include <zmq.h>
}
#include <iostream>
#include <fstream>

bool GenerateCurveKeyPair(std::pair<std::string, std::string> &outKeyPair)
{
    char publicKey[41];
    char secretKey[41];
    int ret = zmq_curve_keypair(publicKey, secretKey);
    if (ret != 0) {
        return true;
    }
    outKeyPair.first = publicKey;
    outKeyPair.second = secretKey;
    return false;
}

static void SaveFile(const std::string &fileName, const std::string &fileContent)
{
    std::cout << "Saving file: " << fileName << std::endl;
    std::ofstream public_file(fileName);
    public_file << fileContent;
    public_file.close();
}

int main(int argc, char **argv)
{
    std::pair<std::string, std::string> keys;
    bool ret = GenerateCurveKeyPair(keys);
    if (ret) {
        std::cout << "create cert for worker failed" << std::endl;
        return -1;
    }

    // 请勿修改文件名
    SaveFile("./worker.key", keys.first);
    SaveFile("./worker.key_secret", keys.second);

    std::pair<std::string, std::string> clientKeys;
    ret = GenerateCurveKeyPair(clientKeys);
    if (ret) {
        std::cout << "create cert for client failed" << std::endl;
        return -1;
    }

    // 请勿修改文件名
    SaveFile("./client.key", clientKeys.first);
    SaveFile("./client.key_secret", clientKeys.second);
}
```

依次执行如下命令：

```shell
g++ create_cert.cpp -lzmq -o create_cert
export LD_LIBRARY_PATH=/usr/local/lib64:/usr/local/lib
./create_cert
```

确认无错误信息打印且目录下生成 `worker.key`、`worker.key_secret`、`client.key` 及 `client.key_secret` 4 个文件。

### 保存 openYuanrong 数据系统组件证书文件

至此已完成 openYuanrong 数据系统组件所需证书文件的生成，执行如下命令，保存这些文件到我们规划好的目录。

```shell
mkdir -p ${WorkSpace}/cert/curve
cd ${WorkSpace}/yr_data_system
cp -ar worker.key worker.key_secret client.key client.key_secret ${WorkSpace}/cert/curve
mkdir -p ${WorkSpace}/cert/curve/worker_authorized_clients
cp ${WorkSpace}/cert/curve/worker.key ${WorkSpace}/cert/curve/worker_authorized_clients/worker.key
cp ${WorkSpace}/cert/curve/client.key ${WorkSpace}/cert/curve/worker_authorized_clients/client.key
```

## 部署 openYuanrong 时配置安全通信

部署主节点参考命令如下，详细配置说明参考[部署参数表](security-params)中的安全配置。

```shell
yr start --master \
--ssl_base_path=${WorkSpace}/cert/yr --ssl_enable=true \
--curve_key_path=${WorkSpace}/cert/curve \
--runtime_ds_encrypt_enable=true --ds_component_auth_enable=true \
--etcd_auth_type=TLS --etcd_ssl_base_path=${WorkSpace}/cert/etcd \
--cache_storage_auth_type=ZMQ --cache_storage_auth_enable=true
```
