# 多语言函数编程接口

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   key_concept.md
   development_guide/index.md
   advanced_tutorials/index.md
   examples/index.md
   api/index
```

多语言函数编程接口提供了扩展分布式应用程序的基本原语：[有状态函数](key-concept-statefull-function)、[无状态函数](key-concept-stateless-function)、[数据对象](key-concept-data-object)和[数据流](key-concept-data-stream)，支持函数以任务或 Serverless 服务方式运行。我们将通过简单示例向您介绍这些核心概念。

## 入门

使用 pip 安装 openYuanrong ，内含 openYuanrong SDK 及命令行工具 yr。

```bash
pip install https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/release/0.6.0/linux/x86_64/openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_x86_64.whl
```

首先导入并初始化 openYuanrong：

```python
import yr

# Init only once
yr.init()
```

:::{admonition} SDK 会自动初始化环境
:class: note

[Driver](#glossary-driver) 中调用 `yr.init()` 接口（不配置 openYuanrong 集群的地址），运行在非 openYuanrong 节点上时，SDK 会尝试拉起一个 openYuanrong 临时环境，该环境在进程退出时**自动销毁**。

:::

## 有状态函数

通过有状态函数可以创建有状态的进程，当调用方法时将保持其内部状态。实例化一个有状态函数时：

1. openYuanrong 将在集群中启动一个专属的进程，有状态函数的方法在该进程中运行，并且可以访问和修改它的状态。

2. 有状态函数将按顺序执行方法调用。

简单示例如下：

```python
# Define stateful function
@yr.instance
class Object:
    def __init__(self):
        self.value = 0

    def save(self, value):
        self.value = value

    def get(self):
        return self.value

# Create a stateful function instance
obj = Object.invoke()

# Asynchronously invoke stateful function, two calls executed in sequence
obj.save.invoke(9)
result_ref = obj.get.invoke()
print(yr.get(result_ref))

# Destroy stateful function instance
obj.terminate()
```

## 无状态函数

无状态函数是在 openYuanrong 集群中并行化 Python、C++、Java 函数的最简单方法。通过以下步骤创建无状态函数：

1. 使用 `@yr.invoke` 装饰您的函数，表示它应该在远端运行。

2. 使用 `.invoke()` 调用函数，返回结果是数据对象的引用。

3. 使用 `yr.get()` 获取数据对象的值。

简单示例如下：

```python
# Define stateless function
@yr.invoke
def say_hello(name):
    return 'hello, ' + name

# Asynchronously invoke stateless functions in parallel
results_ref = [say_hello.invoke('yuanrong') for i in range(3)]

# Retrieve the value of the returned object
print(yr.get(results_ref))
```

## 数据对象

openYuanrong 的分布式共享对象存储可以高效地管理整个集群的数据。在 openYuanrong 中，主要有三种处理数据对象的方式：

1. 隐式创建：无状态函数和有状态函数的返回值会自动存储在 openYuanrong 的分布式对象存储中，只返回一个对象引用。

2. 显式创建：调用 `yr.put()` 将数据对象直接存储。

3. 传递引用：您可以将数据对象引用作为参数传递给其他无状态函数和有状态函数，以避免不必要的数据复制并实现函数延迟执行。

简单示例如下：

```python
# Define stateless function
@yr.invoke
def add(n):
    return n + 1

# Function invoke will return a reference to a data object
result_ref = add.invoke(1)
# The reference to the data object is passed as a parameter
next_result_ref = add.invoke(result_ref)
print(yr.get(next_result_ref))

# Put data to object store
data_ref = yr.put({"key": "value"})

# Get data from object store
print(yr.get(data_ref))  # output {"key": "value"}
```

## 数据流

数据流是可以在多个 openYuanrong 函数间跨节点传递的内存数据，通过 pub/sub 方式访问。数据流在创建生产者或消费者时通过唯一的流名称被隐式创建。

简单示例如下：

```python
# Define stream name
stream_name = "this-stream"

# Create producer, implicitly creating the stream.
producer_config = yr.ProducerConfig(delay_flush_time=5, page_size=1024 * 1024, max_stream_size=1024 * 1024 * 1024, auto_clean_up=True)
producer = yr.create_stream_producer(stream_name, producer_config)
# Produce a piece of data
element = yr.Element(value=b"hello", ele_id=0)
producer.send(element)

# Create consumer and associate it with the stream
consumer_config = yr.SubscriptionConfig("local-consumer")
consumer = yr.create_stream_consumer(stream_name, consumer_config)
# Consume a piece of data
elements = consumer.receive(1000, 1)
```

## 函数服务

您可以将 openYuanrong 函数以 Serverless 服务方式部署，通过 HTTP 请求访问。函数服务定义了函数签名作为请求入口，实现该函数即可以部署为 Serverless 服务。

简单示例如下：

```python
# handler 为函数执行方法入口，每次请求都会触发执行。
# event 为 http 请求传递的数据（Header、Body等）。
# context 为 openYuanrong 提供的运行时上下文，包含函数、执行环境等信息。
def handler(event, context):
    print("received request,event content:", event)

    response = ""
    try:
        response = "hello " + event.get("name")
    except Exception as e:
        print(e)
        response = "please enter your name,for example:{'name':'yuanrong'}"

    return response
```

## 下一步

你可以将 openYuanrong 的简单原语组合起来表达几乎所有的分布式计算模式。想要深入了解 openYuanrong 的[关键概念](./key_concept.md)，请浏览以下用户指南：

- [有状态函数](./development_guide/stateful_function/index.md)
- [无状态函数](./development_guide/stateless_function/index.md)
- [数据对象](./development_guide/data_object/index.md)
- [数据流](./development_guide/data_stream/index.md)
- [函数服务](./development_guide/function_service/index.md)
