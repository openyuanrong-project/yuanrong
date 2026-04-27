yr.config.UserTLSConfig
=======================

.. py:class:: yr.config.UserTLSConfig(root_cert_path: str, module_cert_path: str, module_key_path: str, server_name: str | None = None)

    基类：``object``

    用户在与外部集群通信时使用的 SSL/TLS 配置。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`server_name <UserTLSConfig_server_name>`
         - 服务器名称，默认值为 ``None``。
       * - :ref:`root_cert_path <UserTLSConfig_root_cert_path>`
         - 根证书文件的路径。
       * - :ref:`module_cert_path <UserTLSConfig_module_cert_path>`
         - 模块证书文件的路径。
       * - :ref:`module_key_path <UserTLSConfig_module_key_path>`
         - 模块密钥文件的路径。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <UserTLSConfig_init>`
         -

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.config.UserTLSConfig.__init__
    yr.config.UserTLSConfig.server_name
    yr.config.UserTLSConfig.root_cert_path
    yr.config.UserTLSConfig.module_cert_path
    yr.config.UserTLSConfig.module_key_path