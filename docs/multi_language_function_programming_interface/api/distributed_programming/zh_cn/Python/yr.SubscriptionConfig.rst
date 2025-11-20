yr.SubscriptionConfig
=======================

.. py:class:: yr.SubscriptionConfig(subscription_name: str, subscriptionType: ~yr.stream.SubscriptionType = SubscriptionType.STREAM, extend_config: ~typing.Dict[str, str] = <factory>)

    基类：``object``

    消费者订阅的配置类。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 40 60

       * - :ref:`subscriptionType <subscriptionType>`
         - 订阅类型，包括 ``STREAM``，``ROUND_ROBIN`` 和 ``KEY_PARTITIONS``。
       * - :ref:`subscription_name <subscription_name>`
         - 订阅名，用于标识生产者配置中的订阅。
       * - :ref:`extend_config <extend_config>`
         - 扩展配置。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 40 60

       * - :ref:`__init__ <init_SubscriptionConfig>`
         -

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.SubscriptionConfig.subscriptionType
    yr.SubscriptionConfig.subscription_name
    yr.SubscriptionConfig.extend_config
    yr.SubscriptionConfig.__init__

