# 词汇表

本节将介绍 openYuanrong 常用术语。

{.glossary}

{#glossary-yuanrong-function}
openYuanrong 函数（openYuanrong Function）
: openYuanrong 分布式调度运行的基本单位。相比传统 Serverless 函数概念，openYuanrong 函数更加通用，支持运行中动态创建、长时运行、相互间异步调用、有状态等等，可以表达任意分布式应用的运行实例，起到类似单机 OS 中进程的作用。

{#glossary-stateful-function}
有状态函数（Stateful Function）
: openYuanrong 函数支持有状态，这里的状态是指程序运行中可以访问和修改的进程内私有变量。典型的状态如进程内静态变量、面向对象编程中的成员变量等。基于有状态函数概念，openYuanrong 提供了多语言编程接口，让原来基于 Python、Java 和 C++ 开发的单机类（class）可以自动转换为 openYuanrong 有状态函数运行。

{#glossary-stateless-function}
无状态函数（Stateless Function）
: 无状态函数是有状态函数的特例，它的执行不依赖状态，仅依赖输入参数。基于无状态函数概念，openYuanrong 提供了多语言编程接口，让原来使用 Python、Java 和 C++ 开发的单机函数可以自动转换为 openYuanrong 无状态函数运行。

{#glossary-object}
数据对象（Object）
: 数据对象是可以在多个 openYuanrong 函数间跨节点分布式共享访问的内存数据，支持基于共享内存的高性能 put/get 访问和修改；此外，也可作为函数调用的参数和返回值，自动分布式传递和共享，并支持异步 Future 语义：比如，对某个函数的调用可返回一个数据对象 Future 引用，此时还可将该对象引用作为新的函数调用参数使用；openYuanrong 会在运行中自动解析引用，并通过自动分布式引用计数管理数据对象生命周期。

{#glossary-stream}
数据流（Stream）
: 数据流是可以在多个 openYuanrong 函数间跨节点分布式传递共享的有序无界内存数据集，支持基于共享内存的高性能 pub/sub 访问，支持一对一、一对多、多对一等多种发布订阅模式。通过数据流可方便地解耦多个不同函数，实现多个函数间异步流式数据传递和计算。

{#glossary-master-node}
主节点（Master）
: openYuanrong 集群中包含控制面（集群管理、调度等）和数据面（运行分布式任务）组件的节点。控制面包括的组件有 `function master`，数据面包括的组件有 `function proxy` 、 `function agent` 、 `runtime manager` 和 `data worker` 。

{#glossary-agent-node}
从节点（Agent）
: openYuanrong 集群中只包含数据面（运行分布式任务）组件的节点，包括的组件有 `function proxy` 、 `function agent` 、 `runtime manager` 和 `data worker` 。一台主机上可以部署多个从节点。

{#glossary-functionsystem}
函数系统（Function System）
: openYuanrong 系统之一，提供大规模分布式动态调度，支持函数实例极速弹性扩缩和跨节点迁移，实现集群资源高效利用。

{#glossary-datasystem}
数据系统（Data System）
: openYuanrong 系统之一，提供异构分布式多级缓存，支持 Object、Stream 语义，实现函数实例间高性能数据共享及传递。

{#glossary-driver}
Driver
: 应用的启动进程名称。例如 Python 的启动脚本、C++ 的二进制运行文件或 Java 的可执行 jar 包。

{#glossary-instance}
函数实例（Worker）
: 运行 openYuanrong 函数的进程。

{#glossary-runtime}
运行时（Runtime）
: openYuanrong 函数的运行环境。

{#glossary-process-deployment}
主机部署（进程部署）
: 在主机上以进程方式拉起所有 openYuanrong 组件，并提供基础的健康监测和进程重拉机制。进程部署通常用于轻量级的本地验证及其他对可用性、可靠性、隔离性要求不高的场景。
