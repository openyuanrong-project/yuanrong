# 函数中使用流

本节通过简单的 Python 示例介绍如何在函数中使用流。

## 准备工作

参考[在主机上部署](../../deploy/deploy_processes/index.md)完成openYuanrong部署。

## 在无状态函数中使用

我们在主程序中创建消费者 `local_consumer`，该操作会隐式完成流 `exp-stream` 的创建。生产者为无状态函数，实例在远端运行。生产者和消费者协商使用字符串 `::END::` 作为流结束标志，处理完流后需要主动调用接口 `yr.delete_stream` 删除流，释放资源。

```python
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

        corpus = subprocess.check_output(["python", "-c", "import this"])
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

    # 需要显示删除流，否则流一直存在
    yr.delete_stream(stream_name)
    yr.finalize()
```
