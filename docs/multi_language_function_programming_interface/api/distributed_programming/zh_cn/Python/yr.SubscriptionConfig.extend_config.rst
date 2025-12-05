.. _extend_config_sub:

yr.SubscriptionConfig.extend_config
------------------------------------

.. py:attribute:: SubscriptionConfig.extend_config
   :type: Dict[str, str]

   扩展配置，以字典形式存储，允许用户自定义配置项。
   默认值为空字典，即通过 ``field(default_factory=dict)`` 生成的字典。