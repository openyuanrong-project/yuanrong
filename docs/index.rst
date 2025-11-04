概述
=====

.. raw:: html

   <h4 style="text-align:center">openYuanrong：单机编程体验，分布式运行性能</h4>
   <br>
   <br>

.. grid:: 2
    :gutter: 3 3 4 5

    .. grid-item-card:: 入门
      :link: overview-getting-started
      :link-type: ref
      :text-align: center

    .. grid-item-card:: 安装 openYuanrong
      :link: overview-installation
      :link-type: ref
      :text-align: center

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

.. toctree::
  :hidden:

  overview
  getting_started
  deploy/index.md
  use_cases/index
  multi_language_function_programming_interface/index.md
  more_usage/index.md
  observability/index.md
  FAQ/index.md
  contributor_guide/index.md
  glossary
  security
