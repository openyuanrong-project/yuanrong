# Stream

:::{Note}

头文件介绍：

流缓存主要涉及流对象操作接口及 `Producer`、`Consumer` 两个类，按需要导入下列相关的头文件：

import org.yuanrong.stream.Producer;

import org.yuanrong.stream.ProducerConfig;

import org.yuanrong.stream.Consumer;

import org.yuanrong.stream.SubscriptionConfig;

import org.yuanrong.stream.SubscriptionType;

import org.yuanrong.stream.Element;

:::

## 数据结构

### public class ProducerConfig

创建生产者的配置类。

#### Private Members

```java

private long delayFlushTimeMs = 5L
```

   发送后最多延迟相应时长后触发 flush；<0：代表不自动 flush；为 0：代表立即 flush；否则，表示 delay 的时长，单位为毫秒。

```java

private long pageSizeByte = 1024 * 1024L
```

   代表生产者对应的缓冲 page 大小，单位是 B（字节）；page 写满后会触发 flush。默认 ``1MB``, 必须大于 0 而且是 4KB 的倍数。

```java
private long maxStreamSize = 100 * 1024 * 1024L

```

   指定流在 worker 上最大能使用的共享内存大小，单位是 B（字节）。默认 100MB, 范围 [64KB, worker 共享内存的大小]。

```java
private boolean autoCleanup = false

```

   指定流是否开启自动清理功能。默认 ``false`` 关闭。

```java
private boolean encryptStream = false

```

   指定流是否开启内容加密功能。默认 ``false`` 关闭。

```java
private long retainForNumConsumers = 0L

```

   生产者发送的数据将保留到第 N 个消费者接收。默认值为 ``0`` ，表示生产者发送数据后，如果没有消费者，数据不会保留，等消费者创建后可能会接收不到数据。该参数仅对首次创建的消费者有效，且当前取值范围为 ``[0,1]``,不支持多个消费者。

```java
private long reserveSize = 0L

```

代表预留内存大小，单位是 B（字节）；当创建生产者的时候，会尝试预留 `reserveSize` 字节大小内存，若预留失败，则创建生产者抛出异常。`reserveSize` 必须为 `pageSize` 的整数倍，且取值范围为 ``[0, maxStreamSize]``；若 `reserveSize` 为 ``0``，则会被设置成 `pageSize`。默认值为 ``0``。

```java
private Map<String, String> extendConfig = new HashMap<>()

```

生产者扩展配置。常见的配置项如下：

`"STREAM_MODE"`：流的模式，可选择 ``"MPMC"``、``"MPSC"`` 或者 ``"SPSC"``，默认 ``"MPMC"``，若非上述几种模式则会抛出异常。

其中 ``"MPMC"`` 表示多生产者多消费者，``"MPSC"`` 表示多生产者单消费者，``"SPSC"`` 表示单生产者单消费者。如果是 ``"MPSC"`` 或 ``"SPSC"`` 模式则数据系统内部启用多流共享 Page 功能。

### public class SubscriptionConfig

消费者订阅的配置类。

#### Private Members

```java
private String subscriptionName = ""

```

   订阅名称。

```java
private SubscriptionType subscriptionType = SubscriptionType.STREAM

```

   订阅类型，包括 ``STREAM``、``ROUND_ROBIN`` 和 ``KEY_PARTITIONS`` 三种类型。目前仅支持 ``STREAM`` 类型，其他类型暂时不支持。默认订阅类型为 ``STREAM``。

```java
private Map<String, String> extendConfig = new HashMap<>()

```

   表示扩展配置，为预留字段。

### public class Element

包含 element id 和数据缓存的元素类。

#### Private Members

```java
private long id

```

   Element 的 id。

```java
private ByteBuffer buffer

```

   数据缓存。

## Stream Object Operation Interface

### public static Producer createProducer(String streamName) Throws YRException

创建生产者。

- 参数：

   - **streamName** - 流的名称。长度必须小于 256 个字符且仅含有下列字符 ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。

- 返回：

    Producer：生产者对象。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

### public static Producer createProducer(String streamName, ProducerConfig producerConf) Throws YRException

创建生产者。

- 参数：

   - **streamName** - 流的名称。长度必须小于 256 个字符且仅含有下列字符 ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。
   - **producerConf** - 生产者的配置信息。

- 返回：

    Producer：生产者对象。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

### public static Consumer subscribe(String streamName, SubscriptionConfig config) Throws YRException

创建消费者。

- 参数：

   - **streamName** - 流的名称。长度必须小于 256 个字符且仅含有下列字符 ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。
   - **config** - 消费者的配置信息。

- 返回：

    Consumer：消费者对象。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

### public static Consumer subscribe(String streamName, SubscriptionConfig config, boolean autoAck) Throws YRException

创建消费者。

- 参数：

   - **streamName** - 流的名称。长度必须小于 256 个字符且仅含有下列字符 ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``。
   - **config** - 消费者的配置信息。
   - **autoAck** - 当 `autoAck` = ``true``，消费者在接收到消息时，将为前面的消息自动发送 Ack 给数据系统。

- 返回：

    Consumer：消费者对象。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

## public interface Producer

Producer 接口。

### Interface description

#### void send(Element element) Throws YRException

Producer 发送数据，数据会首先放入缓冲中，根据配置的自动 flush 策略（发送间隔一段时间或者缓冲写满）去刷缓冲或主动调用 flush 让消费者可以访问到。

- 参数：

   - **element** -  需要发送的 Element 数据。Element 可参考公共结构体中 Element 对象结构体。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### void send(Element element, int timeoutMs) Throws YRException

Producer 发送数据，数据会首先放入缓冲中，根据配置的自动 flush 策略 (发送间隔一段时间或者缓冲写满) 去刷缓冲或主动调用 flush 让消费者可以访问到。

- 参数：

   - **element** - 需要发送的 Element 数据。Element 可参考公共结构体中 Element 对象结构体。
   - **timeoutMs** - 超时时间。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### void close() Throws YRException

关闭生产者会触发自动 flush 掉数据缓冲，并且表示不再使用数据缓冲。一旦关闭后，生产者不可再用。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

## public interface Consumer

Consumer 接口。

### Interface description

#### List<Element> receive(long expectNum, int timeoutMs) Throws YRException

消费者接收数据带有订阅功能，接收数据会等待期望个数 `expectNum` 个 `element`，当超时时间 `timeoutMs` 到达或者期望个数的数据可接收到时，该调用返回。

- 参数：

   - **expectNum** - 期望接收 element 的个数。
   - **timeoutMs** - 接收的超时时间。

- 返回：

    List<Element>，一组存放数据的 Element 列表。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### List<Element> receive(int timeoutMs) Throws YRException

消费者接收数据带有订阅功能，接收数据会等待期望个数 `expectNum` 个 `element`，当超时时间 `timeoutMs` 到达或者期望个数的数据可接收到时，该调用返回。

- 参数：

   - **timeoutMs** - 接收的超时时间。

- 返回：

    List<Element>，一组存放数据的 Element 列表。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### void ack(long elementId) Throws YRException

消费者使用完某 `elementId` 标识的 element 后，需要确认已消费完，使得各个 worker 上可以获取到是否所有消费者都已经消费完的信息，若已经消费完某个 Page 可以触发内部的内存回收机制。若不 ack，则在消费者退出时候才会自动 ack。

- 参数：

   - **elementId** - 待确认已消费完成的 element 的 id。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

#### void close() Throws YRException

关闭消费者。一旦关闭后，消费者不可再用。

- 抛出：

   - **YRException** - 统一抛出的异常类型。

## 样例

```java

try {
    ProducerConfig pCfg = ProducerConfig.builder().delayFlushTimeMs(10L).pageSizeByte(2 * 1024 * 1024L).maxStreamSize(200 * 1024 * 1024L).autoCleanup(true).build();
    Producer producer = YR.createProducer("aaaaaaaa", pCfg);

    SubscriptionConfig config = SubscriptionConfig.builder().subscriptionName("aaa").build();
    Consumer consumer = YR.subscribe("aaaaaaaa", config);

    String toSend = args;
    ByteBuffer buffer = ByteBuffer.wrap(toSend.getBytes());
    Element element = new Element(111L, buffer);
    producer.send(element);

    List<Element> recv = consumer.receive(3, 6000);
    if (recv.isEmpty()) {
        // 处理空值。
    }
    Element e = recv.get(0);
    Charset charset = Charset.forName("UTF-8");
    String res = charset.decode(e.getBuffer()).toString();
    consumer.ack(e.getId());

    producer.close();
    consumer.close();
    YR.deleteStream("aaaaaaaa");
} catch (YRException e) {
    // 处理异常。
}
```
