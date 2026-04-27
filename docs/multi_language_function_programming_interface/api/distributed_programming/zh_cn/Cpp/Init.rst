Init
======

.. cpp:function:: ClientInfo YR::Init(const Config &conf)

    openYuanrong 初始化 API，用于配置运行模式和系统参数。参数规范请参考 :doc:`结构体说明 <./struct-Config>`。

    .. code-block:: cpp

       // 本地模式
       YR::Config conf;
       conf.mode = YR::Config::Mode::LOCAL_MODE;
       conf.threadPoolSize = 10;
       YR::Init(conf);
    
    .. code-block:: cpp

       // 集群模式
       YR::Config conf;
       conf.mode = YR::Config::Mode::LOCAL_MODE;
       conf.threadPoolSize = 10;
       YR::Init(conf);

    .. note::
        当 openYuanrong 集群启用多租户功能时，用户需配置租户ID。有关租户ID配置的详细信息，请参阅 :doc:`结构体说明 <./struct-Config>` 中关于“租户 id”的章节。

    参数：
        - **conf** - openYuanrong 初始化参数配置。参数规范请参考 :doc:`结构体说明 <./struct-Config>`。
    
    抛出：
        :cpp:class:`Exception` - 当检测到无效的配置参数（例如无效的模式类型）时，系统将抛出异常。

    返回：
        客户端信息（ClientInfo）的详细信息请参考 :doc:`结构体说明 <./struct-Config>`。