yr.Alarm
=========

.. py:class:: yr.Alarm(name: str, description: str)

    基类：``Metrics``

    用于设置和管理告警信息。

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
        - **name** (str) - 名称。
        - **description** (str) - 描述。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_Alarm>`
         -
       * - :ref:`set <Alarm_set>`
         - 设置值。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Alarm.__init__
    yr.Alarm.set
