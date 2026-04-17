yr.UInt64Counter
==================

.. py:class:: yr.UInt64Counter(name: str, description: str, unit: str, labels: ~typing.Dict[str, str] = Field(name=None, type=None, default=<dataclasses._MISSING_TYPE object>, default_factory=<class 'dict'>, init=True, repr=True, hash=None, compare=True, metadata=mappingproxy({}), _field_type=None))

    基类：``Metrics``

    表示 64 位无符号整数计数器的类。

    .. note::
        使用本样例前，需在部署时配置 `ENABLE_METRICS` 和 `RUNTIME_METRICS_CONFIG` 环境变量，否则样例无法使用。可通过以下方式配置：

        方式一：在 config.toml 中配置

        .. code-block:: toml

            [function_agent.env]
            ENABLE_METRICS = "true"
            RUNTIME_METRICS_CONFIG = '{"backends":[{"immediatelyExport":{"name":"FileExporter","enable":true,"exporters":[{"fileExporter":{"enable":true,"enabledInstruments":["yr_alarm"],"failureQueueMaxSize":1000,"failureDataDir":"/home/sn/metrics/failure","failureDataFileMaxCapacity":20,"initConfig":{"fileDir":"/home/sn/metrics/file","rolling":{"enable":true,"maxFiles":3,"maxSize":100,"compress":false},"contentType":"STANDARD"}}}]}}]}'

        方式二：通过命令行参数覆盖

        .. code-block:: bash

            yr start --master -s 'function_agent.env.ENABLE_METRICS="true"' -s 'function_agent.env.RUNTIME_METRICS_CONFIG="{\"backends\":[{\"immediatelyExport\":{\"name\":\"FileExporter\",\"enable\":true,\"exporters\":[{\"fileExporter\":{\"enable\":true,\"enabledInstruments\":[\"yr_alarm\"],\"failureQueueMaxSize\":1000,\"failureDataDir\":\"/home/sn/metrics/failure\",\"failureDataFileMaxCapacity\":20,\"initConfig\":{\"fileDir\":\"/home/sn/metrics/file\",\"rolling\":{\"enable\":true,\"maxFiles\":3,\"maxSize\":100,\"compress\":false},\"contentType\":\"STANDARD\"}}}]}}]}"'

    参数：
        - **name** (str) - 计数器名称。
        - **description** (str) - 计数器描述。
        - **unit** (str，可选) - 单位。
        - **labels** (Dict[str, str]，可选) - 计数器的可选标签。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_UInt64Counter>`
         -
       * - :ref:`add_labels <add_labels_UInt64Counter>`
         - 为 metrics 数据添加标签。
       * - :ref:`set <set_UInt64Counter>`
         - 将 uint64 计数器设置为给定的值。
       * - :ref:`reset <reset_UInt64Counter>`
         - 重置 uint64 计数器。
       * - :ref:`increase <increase_UInt64Counter>`
         - 将 uint64 计数器增加到给定的值。
       * - :ref:`get_value <get_value_UInt64Counter>`
         - 获取 uint64 计数器的值。



.. toctree::
    :maxdepth: 1
    :hidden:

    yr.UInt64Counter.__init__
    yr.UInt64Counter.add_labels
    yr.UInt64Counter.set
    yr.UInt64Counter.reset
    yr.UInt64Counter.increase
    yr.UInt64Counter.get_value
