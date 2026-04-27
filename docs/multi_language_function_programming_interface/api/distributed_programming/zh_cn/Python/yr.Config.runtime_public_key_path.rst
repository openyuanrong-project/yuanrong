.. _runtime_public_key_path:

yr.Config.runtime_public_key_path
------------------------------------

.. py:attribute:: config.runtime_public_key_path
   :type: str
   :value: ''

   客户端公钥路径，用于数据系统 TLS 认证。如果启用了数据系统加密（`enable_ds_encrypt` 为 `true`）且运行时公钥路径（`runtime_public_key_path`）为空，则会抛出异常。