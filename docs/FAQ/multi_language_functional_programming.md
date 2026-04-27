# 多语言函数编程 FAQ

## 一、无状态函数、有状态函数问题

### Driver 中正确配置了 http_proxy/https_proxy 代理地址，但  `yr.init` 接口连接 openYuanrong 集群失败，报错：“ValueError: failed to init, code: 4002, module code 20, msg: failed to connect to all addresses”

原因：openYuanrong 默认不生效 http_proxy/https_proxy 代理配置。

解决方法：配置如下环境变量，生效 http_proxy/https_proxy 代理配置。

  ```shell
  export YR_ENABLE_HTTP_PROXY=true
  ```
