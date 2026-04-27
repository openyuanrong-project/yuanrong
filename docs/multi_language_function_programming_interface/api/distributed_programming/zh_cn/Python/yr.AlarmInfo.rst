yr.AlarmInfo
=======================

.. py:class:: yr.AlarmInfo(alarm_name: str = '', alarm_severity: yr.runtime.AlarmSeverity = AlarmSeverity.OFF, location_info: str = '', cause: str = '', starts_at: int = -1, ends_at: int = -1, timeout: int = -1, custom_options: typing.Dict[str, str])

    基类：``object``

    告警信息类。

    **属性**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`alarm_name <alarm_name>`
         - 告警的名称。
       * - :ref:`alarm_severity <alarm_severity>`
         - 告警的严重性，类型为 ``AlarmSeverity``，默认值为 ``AlarmSeverity.OFF``。
       * - :ref:`cause <cause>`
         - 告警原因的描述。默认值为空字符串。
       * - :ref:`ends_at <ends_at>`
         - 告警结束的时间戳。默认值为 ``-1``。
       * - :ref:`location_info <location_info>`
         - 告警的位置信息。
       * - :ref:`starts_at <starts_at>`
         - 告警开始的时间戳。默认值为 ``-1``。
       * - :ref:`timeout <timeout_AI>`
         - 告警的超时持续时间，默认值为 ``-1``。
       * - :ref:`custom_options <custom_options>`
         - 自定义选项，存储为键值对字典，默认值为空字典。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_AlarmInfo>`
         -


.. toctree::
    :maxdepth: 1
    :hidden:

    yr.AlarmInfo.__init__
    yr.AlarmInfo.alarm_name
    yr.AlarmInfo.alarm_severity
    yr.AlarmInfo.cause
    yr.AlarmInfo.ends_at
    yr.AlarmInfo.location_info
    yr.AlarmInfo.starts_at
    yr.AlarmInfo.timeout
    yr.AlarmInfo.custom_options
