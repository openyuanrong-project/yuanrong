yr.Gauge
=========

.. py:class:: yr.Gauge(name: str, description: str, unit: str = '', labels: Dict[str, str] = {})

    基类：``Metrics``

    用来上报 Metrics 数据。

    .. note::
        1. 类似 Prometheus 的数据结构。
        2. 计费信息会上报到后台的 openYuanrong 采集器。
        3. 不能在 driver 中使用。
        4. 使用本样例前，需在部署时配置 `ENABLE_METRICS` 和 `RUNTIME_METRICS_CONFIG` 环境变量，否则样例无法使用。可通过以下方式配置：

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
        - **unit** (str，可选) - 单位。
        - **labels** (Dict[str, str]，可选) - 标签。例如，``labels = {"request_id": "abc"}``。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_Gauge>`
         -
       * - :ref:`add_labels <add_labels>`
         - 添加 label。
       * - :ref:`set <set_Gauge>`
         - 设置值。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Gauge.__init__
    yr.Gauge.add_labels
    yr.Gauge.set
