# 关键概念

(key-concept-statefull-function)=

## 有状态函数

[openYuanrong 函数](glossary-yuanrong-function)支持有状态，这里的状态是指程序运行中可以访问和修改的进程内私有变量。典型的状态如进程内静态变量、面向对象编程中的成员变量等。基于有状态函数概念，openYuanrong 提供了多语言编程接口，让原来基于 Python、Java 和 C++ 开发的单机类（class）可以自动转换为 openYuanrong 有状态函数运行。

查看[有状态函数开发指南](./development_guide/stateful_function/index.md)。

(key-concept-stateless-function)=

## 无状态函数

无状态函数是有状态函数的特例，它的执行不依赖状态，仅依赖输入参数。基于无状态函数概念，openYuanrong 提供了多语言编程接口，让原来使用 Python、Java 和 C++ 开发的单机函数可以自动转换为 openYuanrong 无状态函数运行。

查看[无状态函数开发指南](./development_guide/stateless_function/index.md)。

(key-concept-data-object)=

## 数据对象

数据对象是可以在多个 openYuanrong 函数间跨节点分布式共享访问的内存数据，支持基于共享内存的高性能 put/get 访问和修改；此外，也可作为函数调用的参数和返回值，自动分布式传递和共享，并支持异步 Future 语义：比如，对某个函数的调用可返回一个数据对象 Future 引用，此时还可将该对象引用作为新的函数调用参数使用；openYuanrong 会在运行中自动解析引用，并通过自动分布式引用计数管理数据对象生命周期。

查看[数据对象开发指南](./development_guide/data_object/index.md)。

(key-concept-data-stream)=

## 数据流

数据流是可以在多个 openYuanrong 函数间跨节点分布式传递共享的有序无界内存数据集，支持基于共享内存的高性能 pub/sub 访问，支持一对一、一对多、多对一等多种发布订阅模式。通过数据流可方便地解耦多个不同函数，实现多个函数间异步流式数据传递和计算。

查看[数据流开发指南](./development_guide/data_stream/index.md)。
