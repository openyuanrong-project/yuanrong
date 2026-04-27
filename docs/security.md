# 安全

openYuanrong 组件需要在集群内通信以完成请求转发、任务调度等功能，组件间通信可使用 TLS 加密以保证安全。

使用 openYuanrong 开发的应用，通过客户端连接到 openYuanrong 集群运行。openYuanrong 允许客户端运行任意代码，并对集群及底层计算资源有完全的访问权限。

## 核心准则

在 openYuanrong 生态中，主要有三类角色：开发者通过 openYuanrong API 构建应用程序，并在本地或基础设施提供商提供的集群环境中进行应用的测试与部署；基础设施服务商为开发者提供运行 openYuanrong 的计算环境；用户使用基于 openYuanrong 平台开发的各类应用程序。

为确保 openYuanrong 集群的安全运行，所有参与者需共同遵循以下关键要求：

- 网络边界防护：所有集群部署必须实施有效的网络隔离策略，确保集群外部环境的安全防护措施到位。

- 组件通信安全：基础设施服务商需建立受控的隔离网络环境，确保 openYuanrong 各组件的通信安全。当集群需要访问第三方服务时，必须实施严格的身份验证机制和访问控制策略。

- 代码可信管理：openYuanrong 平台本身不执行代码安全审查，开发者需对部署的应用程序代码进行完整的安全验证和风险评估。

- 租户和负载隔离：当存在多租户或工作负载隔离需求时，基础设施服务商应根据开发者的具体要求，部署独立且相互隔离的 openYuanrong 集群环境。

总之，所有操作均需在经过安全认证的网络环境中执行，确保平台仅处理经过验证的可信代码。基础设施服务商和开发方应建立联合安全机制，通过网络隔离、访问控制和代码审计等多层防护，构建完整的安全防护体系。

## 漏洞管理

openYuanrong 在 openEuler 社区开源，如果您发现了 openYuanrong 的安全问题或疑似安全漏洞，请遵循 openEuler 社区提供的[漏洞管理](https://www.openeuler.org/zh/security/vulnerability-reporting/){target="_blank"}处理机制上报给安全委员会。我们会快速的响应、分析和解决。

## 注意事项

1. 通过限制权限等措施确保访问 openYuanrong 集群的人员是受信任的。

2. openYuanrong 使用 cloudpickle 来序列化 Python 对象。请参阅 [pickle 文档](https://docs.python.org/3/library/pickle.html){target="_blank"}了解更多安全模型信息。

3. openYuanrong 组件在运行时已提供对证书格式、有效期、信任链及数字签名的核心校验。如果您有更高安全需求，建议基于现有校验框架，参考行业安全实践（如 OpenSSL）进行拓展。

4. 当前版本包含了多租户特性的早期实现。该特性尚处于开发阶段，未达到生产标准，API 与架构可能在后续版本中发生重大调整，正式发布计划请关注我们的 Roadmap。
