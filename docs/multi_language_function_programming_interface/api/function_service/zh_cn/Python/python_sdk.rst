Python SDK
================

yr.Context
---------------------

.. py:class:: yr.Context(options: dict)

    基类：``object``

    openYuanrong 运行时提供的上下文信息。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_context>`
         -
       * - :ref:`getCPUNumber <getCPUNumber>`
         - 获取分配给正在运行的函数的 CPU 数量（CPU 数量按千分之一核计算，1 个 CPU 核心等于 1000 千分之一核）。
       * - :ref:`getFunctionName <getFunctionName>`
         - 获取函数名。
       * - :ref:`getLogger <getLogger>`
         - 获取用于用户在标准输出中打印日志的记录器，SDK 中必须提供 Logger 接口。
       * - :ref:`getMemorySize <getMemorySize>`
         - 获取分配给正在运行的函数的内存大小。
       * - :ref:`getPackage <getPackage>`
         - 获取函数包。
       * - :ref:`getRequestID <getRequestID>`
         - 获取 Request ID。
       * - :ref:`getTenantID <getTenantID>`
         - 获取租户 ID。
       * - :ref:`getUserData <getUserData>`
         - 根据键获取用户通过环境变量传入的值。
       * - :ref:`getVersion <getVersion>`
         - 获取函数版本。
       * - :ref:`get_session_id <get_session_id>`
         - 获取当前的 session ID。
       * - :ref:`get_session_service <get_session_service>`
         - 获取 SessionService 实例，用于加载和修改当前调用的 Session 对象。
       * - :ref:`set_session_id <set_session_id>`
         - 设置当前调用的 session ID。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Context.__init__
    yr.Context.getCPUNumber
    yr.Context.getFunctionName
    yr.Context.getLogger
    yr.Context.getMemorySize
    yr.Context.getPackage
    yr.Context.getRequestID
    yr.Context.getTenantID
    yr.Context.getUserData
    yr.Context.getVersion
    yr.Context.get_session_service
    yr.Context.set_session_id

yr.SessionService
---------------------

.. py:class:: yr.SessionService()

    提供会话加载能力的 SDK 接口。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`load_session <load_session>`
         - 加载当前调用关联的会话对象。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.SessionService.load_session

yr.Session
---------------------

.. py:class:: yr.Session()

    表示一个 Agent 会话对象。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`wait_for_notify <wait_for_notify>`
         - 阻塞当前执行并等待输入。
       * - :ref:`notify <notify>`
         - 唤醒正在等待的线程。
       * - :ref:`is_interrupted <is_interrupted>`
         - 检查当前会话是否已被外部中断。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Session.wait_for_notify
    yr.Session.notify
    yr.Session.is_interrupted

yr.Function
------------------

.. py:class:: yr.Function(function_name: str, context_: Context | None = None)

    基类：``object``

    提供函数互调能力。

    **方法**：

    .. list-table::
       :header-rows: 0
       :widths: 30 70

       * - :ref:`__init__ <init_func>`
         -
       * - :ref:`invoke <invoke_func>`
         - 调用函数。
       * - :ref:`options <options_func>`
         - 设置用户调用选项。

.. toctree::
    :maxdepth: 1
    :hidden:

    yr.Function.__init__
    yr.Function.invoke
    yr.Function.options
    
