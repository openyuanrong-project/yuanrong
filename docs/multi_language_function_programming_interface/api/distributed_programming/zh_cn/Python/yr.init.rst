yr.init
=====================

.. py:function:: yr.init(conf: Config | None = None) -> ClientInfo

    初始化客户端，根据配置连接到 openYuanrong 集群。 

    参数：
        - **conf** (Config_，可选) – openYuanrong 的初始化参数配置。参数说明详见 `Config` 数据结构。该参数可选，为空时从环境变量导入。
    
    返回：
        此次调用的上下文信息。
        数据类型：ClientInfo_ 。

    异常：
        - **RuntimeError** – 如果 yr.init 被调用两次。
        - **TypeError** – 如果参数类型错误。
        - **ValueError** – 如果参数值错误。
	
    样例：
        >>> import yr
        >>>
        >>> conf = yr.Config()
        >>> yr.init(conf)

.. _Config: ../../zh_cn/Python/yr.Config.html#yr.Config
.. _ClientInfo: ../../zh_cn/Python/yr.config.ClientInfo.html#yr.config.ClientInfo
