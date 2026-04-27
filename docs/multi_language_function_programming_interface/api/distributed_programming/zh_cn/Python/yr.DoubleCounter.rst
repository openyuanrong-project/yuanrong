yr.DoubleCounter
==================

.. py:class:: yr.DoubleCounter(name: str, description: str, unit: str, labels: ~typing.Dict[str, str] = Field(name=None, type=None, default=<dataclasses._MISSING_TYPE object>, default_factory=<class 'dict'>, init=True, repr=True, hash=None, compare=True, metadata=mappingproxy({}), _field_type=None))

    基类：``Metrics``

    表示双精度计数器类。

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
        - **labels** (Dict[str, str]，可选) - 计数器的标签，默认为空字典。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_DoubleCounter>`
         -
       * - :ref:`add_labels <add_labels_DoubleCounter>`
         - 为 metrics 数据添加标签。
       * - :ref:`set <set_DoubleCounter>`
         - 将双精度计数器设置为给定的值。
       * - :ref:`reset <reset_DoubleCounter>`
         - 重置双精度计数器。
       * - :ref:`increase <increase_DoubleCounter>`
         - 将双精度计数器增加到给定的值。
       * - :ref:`get_value <get_value_DoubleCounter>`
         - 获取双精度计数器的值。



.. toctree::
    :maxdepth: 1
    :hidden:

    yr.DoubleCounter.__init__
    yr.DoubleCounter.add_labels
    yr.DoubleCounter.set
    yr.DoubleCounter.reset
    yr.DoubleCounter.increase
    yr.DoubleCounter.get_value
