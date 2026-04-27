.. _ds_public_key_path:

yr.Config.ds_public_key_path
------------------------------------

.. py:attribute:: config.ds_public_key_path
   :type: str
   :value: ''

   工作进程公钥路径，用于数据系统 TLS 认证。如果启用了数据系统加密（`enable_ds_encrypt` 为 ``true``）且数据系统公钥路径（`ds_public_key_path`）为空，则会抛出异常。