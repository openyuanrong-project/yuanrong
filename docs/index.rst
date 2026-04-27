.. title:: 首页

.. raw:: html

    <h2 style="text-align:center">openYuanrong</h2>
    <h6 style="text-align:center">单机编程体验，分布式运行性能</h6>
    
    <hr style="border:none; border-top:1px solid #000; margin:20px 0; width:100%;">
    
    <p style="text-align:left; margin:20px 0; line-height:1.6;">
        openYuanrong 是一个 Serverless 分布式计算引擎，致力于以一套统一的 Serverless 架构支持 AI、大数据、微服务等各类分布式应用。它提供多语言函数编程接口，以单机编程体验简化分布式应用开发；提供分布式动态调度和数据共享等能力，实现分布式应用的高性能运行和集群的高效资源利用。
    </p>

.. raw:: html

    <div style="display: flex; justify-content: center; gap: 20px; flex-wrap: wrap; margin: 20px 0;">

.. button-ref:: overview-getting-started
    :ref-type: ref
    :color: primary
    :shadow:

    快速入门

.. button-ref:: overview-installation
    :ref-type: ref
    :color: primary
    :shadow:

    安装部署

.. button-link:: https://www.openeuler.openatom.cn/zh/sig/sig-YuanRong#repositories  
    :color: primary
    :shadow:

    源码仓列表

.. raw:: html

    </div>


.. grid:: 3
    :gutter: 3 3 4 5

    .. grid-item-card::

        **了解使用案例**
        ^^^
        openYuanrong 可作为 AI 智算的基础设施，用于开发 AI 推理、强化学习等应用，也可作为通用计算的基础设施，开发大数据分析、HPC 等应用。
        +++
        :doc:`使用案例 <./use_cases/index>`

    .. grid-item-card::

        **开发分布式应用**
        ^^^
        openYuanrong 以函数为开发粒度，提供多语言函数编程接口，快速构建和运行分布式应用。
        +++
        :doc:`多语言函数编程接口 <./multi_language_function_programming_interface/index>`

    .. grid-item-card::

        **部署工作负载**
        ^^^
        openYuanrong 支持在单机上部署和运行，用于学习和开发，生产时可无缝扩展到大型集群。
        +++
        :doc:`部署 openYuanrong <./deploy/index>`

.. raw:: html

    <h2 style="text-align:center">动态</h2>


.. grid:: 2
    :gutter: 3 3 4 5

    .. grid-item-card::

        **博客**
        ^^^
        - `把集群变“单机”（下）——openYuanrong核心架构设计解析 <https://www.openeuler.openatom.cn/zh/blog/20260131-openYuanrong_02/20260131-openYuanrong_02.html>`_
        - `把集群变“单机”（上）——openYuanrong核心技术理念解析 <https://www.openeuler.openatom.cn/zh/blog/20260128-openYuanrong/20260128-openYuanrong.html>`_
        
        `openEuler 官网博客 <https://www.openeuler.openatom.cn/zh/interaction/blog-list/>`_ 标签栏搜索 openYuanrong 查看更多。

    .. grid-item-card::

        **新版本发布**
        ^^^
        v0.7.0 已发布，新增支持 k8s 部署和 spring 微服务兼容，以及分布式调度/缓存等功能。       
        `Release note & Package <https://gitcode.com/openeuler/yuanrong/releases/v0.7.0>`_ 查看更多。

.. raw:: html

    <h2 style="text-align:center">加入 openYuanrong</h2>
    
    <div style="display: flex; justify-content: center; margin: 20px 0;">
        <div style="border: 1px solid #ddd; border-radius: 8px; padding: 20px; max-width: 800px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
            <h3 style="margin-top: 0; border-bottom: 1px solid #eee; padding-bottom: 10px;">会议与活动</h3>
            <div style="display: flex; gap: 20px; flex-wrap: wrap;">
                <div style="flex: 1; min-width: 250px;">
                    <p>参与 openYuanrong SIG 组会议：</p>
                    <ul>
                        <li><a href="https://mailweb.openeuler.org/postorius/lists/yuanrong@openeuler.org/ ">订阅邮件(yuanrong@openeuler.org)</a></li>
                        <li><a href="https://www.openeuler.openatom.cn/zh/sig/sig-YuanRong ">例会信息</a></li>
                    </ul>
                </div>
                <div style="flex: 1; min-width: 150px; text-align: center;">
                    <img src="_images/qrcode.png" style="width: 150px;">
                </div>
            </div>
        </div>
    </div>
    
.. toctree::
  :hidden:

  getting_started
  deploy/index.md
  use_cases/index
  multi_language_function_programming_interface/index.md
  more_usage/index.md
  observability/index.md
  FAQ/index.md
  contributor_guide/index.md
  reference/index.md
  security