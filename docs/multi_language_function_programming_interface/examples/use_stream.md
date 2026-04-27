# 函数中使用数据流

本节通过简单的 Python 示例介绍如何在函数中使用流。

## 准备工作

参考[在主机上部署](../../deploy/deploy_processes/index.md)完成openYuanrong部署。

## 在无状态函数中使用流

我们在主程序中创建消费者 `local_consumer`，该操作会隐式完成流 `exp-stream` 的创建。生产者为无状态函数，实例在远端运行。生产者和消费者协商使用字符串 `::END::` 作为流结束标志，处理完流后需要主动调用接口 `yr.delete_stream` 删除流，释放资源。

```python
import sys
import subprocess
import yr
import time

@yr.invoke
def send_stream(stream_name, end_marker):
    try:
        # 创建生产者，配置自动 ACK
        # 流发送会进行缓存，对于实时性要求较高的任务，可调低 delay_flush_time 的值，默认 5ms
        producer_config = yr.ProducerConfig(delay_flush_time=5, page_size=1024 * 1024, max_stream_size=1024 * 1024 * 1024, auto_clean_up=True)
        stream_producer = yr.create_stream_producer(stream_name, producer_config)

        corpus = subprocess.check_output([sys.executable, "-c", "import this"])
        lines = corpus.decode().split("\n")

        i = 0
        for line in lines:
            if len(line) > 0:
                # 发送流
                stream_producer.send(yr.Element(line.encode(), i))
                print("send:" + line)
                i += 1

        # 发送业务约定的结束符号，关闭生产者
        stream_producer.send(yr.Element(end_marker.encode(), i))
        stream_producer.close()
        print("stream producer is closed")
    except RuntimeError as exp:
        print("unexpected exp: ", exp)


if __name__ == '__main__':
    yr.init()

    stream_name = "exp-stream"
    end_marker = "::END::"
    # 创建消费者，隐式创建流
    config = yr.SubscriptionConfig("local_consumer")
    consumer = yr.create_stream_consumer(stream_name, config)
    send_stream.invoke(stream_name, end_marker)

    end = False
    while not end:
        # 经过 1000ms 或收到 10 个 elements 就返回
        elements = consumer.receive(1000, 10)
        for e in elements:
            data_str = e.data.decode()
            print("receive:" + data_str)
            # 收到约定的结束符后，关闭消费者
            if data_str == end_marker:
                consumer.close()
                print("stream consumer is closed")
                end = True

    yr.finalize()
```

## 在函数服务中使用流

我们创建一个函数服务作为流的生产者，一个 http 客户端作为消费者接收流。通过 REST API 订阅流服务时，需要先触发订阅，再执行流生产。

### 注册生产者函数

```python
# producer.py
import sys
import subprocess
import yr

def handler(event, context):
    print("received request,event content:", event)

    try:
        # 读取请求参数中流名称和流结束标记
        stream_name = event.get("stream_name")
        stream_end_marker = event.get("stream_end_marker")
        # 创建生产者，配置自动 ACK
        # 流发送会进行缓存，对于实时性要求较高的任务，可调低 delay_flush_time 的值，默认 5ms
        producer_config = yr.ProducerConfig(delay_flush_time=5, page_size=1024 * 1024, max_stream_size=1024 * 1024 * 1024, auto_clean_up=True)
        stream_producer = yr.create_stream_producer(stream_name, producer_config)

        corpus = subprocess.check_output([sys.executable, "-c", "import this"])
        lines = corpus.decode().split("\n")

        i = 0
        for line in lines:
            if len(line) > 0:
                stream_producer.send(yr.Element(line.encode(), i))
                print("send:" + line)
                i += 1

        # 发送业务约定的结束符号，关闭生产者
        stream_producer.send(yr.Element(stream_end_marker.encode(), i))
        stream_producer.close()
        print("stream producer is closed")
    except RuntimeError as e:
        print(e)
        return "failed, yuanrong runtime error"
    except Exception as e:
        print(e)
        return "failed,request body format:{'stream_name':'this-stream','stream_end_marker':'::END::'}"

    return "ok"
```

使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

```bash
# 替换 /opt/mycode/service 为您 producer.py 代码所在目录
META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 IP}:31182>
curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0@myService@stream-producer","runtime":"python3.9","handler":"producer.handler","kind":"faas","cpu":600,"memory":512,"timeout":60,"storageType":"local","codePath":"/opt/mycode/service"}'
```

结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:default:function:0@myService@stream-producer:latest`。

```bash
{"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@myService@stream-producer:latest","createTime":"2025-06-28 08:57:06.856 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@myService@stream-producer","name":"0@myService@stream-producer","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@myService@stream-producer:latest","revisionId":"20250628085706856","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"producer.handler","layers":null,"cpu":600,"memory":512,"runtime":"python3.9","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-06-28 08:57:06.856 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
```

### 启动消费者进程

运行如下程序，调用[订阅流服务 API](../api/function_service/stream_invocation.md) 订阅流 `this-stream`，参考[获取 frontend 组件服务端点](api-frontend-endpoint)替换代码中的 `{frontend_endpoint}`。

```python
# consumer.py
import requests

url = 'http://{frontend_endpoint}/serverless/v1/stream/subscribe'
headers = {
    'X-Stream-Name': 'this-stream',
    'X-Expect-Num': '10',
    'X-Timeout-Ms': '5000'
}

# 发送请求并启用流式响应
try:
    response = requests.get(url, headers=headers, stream=True)

    if response.status_code == 200:
        for chunk in response.iter_content(chunk_size=1024):
            if chunk:
                print(chunk.decode("utf-8"))
        print('successed')
    else:
        print(f'subscribe failed: http_code={response.status_code}, body={response.text}')
except requests.RequestException as e:
    print(f"request failed: {e}")
```

### 运行生产者函数

使用 curl 工具调用生产者服务，创建流 `this-stream` 并生产内容, API 参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

```bash
FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888`>
FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"stream_name":"this-stream","stream_end_marker":"::END::"}'
```

结果输出：

```bash
HTTP/1.1 200 OK
Content-Type: application/json
X-Billing-Duration: this is billing duration TODO
X-Inner-Code: 0
X-Invoke-Summary:
X-Log-Result: dGhpcyBpcyB1c2VyIGxvZyBUT0RP
Date: Sat, 28 Jun 2025 08:59:28 GMT
Content-Length: 4

"ok"
```

此时，消费者进程输出如下：

```bash
The Zen of Python, by Tim PetersBeautiful is better than ugly.Explicit is better than implicit.Simple is better than complex.Complex is better than complicated.Flat is better than nested.Sparse is better than dense.Readability counts.Special cases aren't special enough to break the rules.Although practicality beats purity.

Errors should never pass silently.Unless explicitly silenced.In the face of ambiguity, refuse the temptation to guess.There should be one-- and preferably only one --obvious way to do it.Although that way may not be obvious at first unless you're Dutch.Now is better than never.Although never is often better than *right* now.If the implementation is hard to explain, it's a bad idea.If the implementation is easy to explain, it may be a good idea.Namespaces are one honking great idea -- let's do more of those!

::END::

successed
```
