# Stream

:::{Note}

Introduction to header files:

Stream caching primarily involves the stream object operation interfaces and the `Producer` and `Consumer` interfaces. Import the following relevant header files as needed:

import org.yuanrong.stream.Producer;

import org.yuanrong.stream.ProducerConfig;

import org.yuanrong.stream.Consumer;

import org.yuanrong.stream.SubscriptionConfig;

import org.yuanrong.stream.SubscriptionType;

import org.yuanrong.stream.Element;

:::

## Data Structure description

### public class ProducerConfig

Configuration class for creating the producer.

#### Private Members

```java

private long delayFlushTimeMs = 5L
```

After sending, the flush is triggered after the maximum delay time. <0: No automatic flush. 0: Flush immediately. Otherwise, it indicates the delay time in milliseconds.

```java

private long pageSizeByte = 1024 * 1024L
```

   Represents the buffer page size corresponding to the producer, in bytes (B); when the page is full, flush is triggered.

   The default is ``1MB``, and it must be greater than 0 and a multiple of 4KB.

```java
private long maxStreamSize = 100 * 1024 * 1024L

```

   Specifies the maximum amount of shared memory that a stream can use on a worker, in units of B (bytes).

   The default is ``100MB``, with a range of [64KB, the size of the worker’s shared memory].

```java
private boolean autoCleanup = false

```

   Specifies whether the stream has the automatic cleanup feature enabled.

   The default is ``false``, which means it is disabled.

```java
private boolean encryptStream = false

```

   Specifies whether the stream has the content encryption feature enabled.

   The default is ``false``, which means it is disabled.

```java
private long retainForNumConsumers = 0L

```

   The data sent by the producer will be retained until the Nth consumer receives it.

   The default value is ``0``, which means that if there are no consumers when the producer sends the data, the data will not be retained, and the consumer might not receive the data after it is created. This parameter is only effective for the first consumer created, and the current valid range is ``[0, 1]``, and it does not support multiple consumers.

```java
private long reserveSize = 0L

```

   Represents the reserved memory size, in units of B (bytes).

   When creating a producer, it will attempt to reserve reserveSize bytes of memory. If the reservation fails, an exception will be thrown during the creation of the producer. `reserveSize` must be an integer multiple of `pageSize`, and its value range is ``[0, maxStreamSize]``. If `reserveSize` is ``0``, it will be set to `pageSize`. The default value is ``0``.

```java
private Map<String, String> extendConfig = new HashMap<>()

```

   Producer expansion configuration. Common configuration items are as follows:

   `"STREAM_MODE"`: Stream mode, can be ``"MPMC"``, ``"MPSC"``, or ``"SPSC"``, default is ``"MPMC"``. If it is not one of the above modes, an exception will be thrown.

   ``"MPMC"`` stands for multiple producers and multiple consumers; ``"MPSC"`` stands for multiple producers and single consumer; ``"SPSC"`` stands for single producer and single consumer.

   If it is ``"MPSC"`` or ``"SPSC"`` mode, the data system internally enables the multi-stream shared Page function.

### public class SubscriptionConfig

Consumer subscription configuration class.

#### Private Members

```java
private String subscriptionName = ""

```

   Subscription name.

```java
private SubscriptionType subscriptionType = SubscriptionType.STREAM

```

   Subscription types include ``STREAM``, ``ROUND_ROBIN``, and ``KEY_PARTITIONS``.

   Currently, only the ``STREAM`` type is supported. Other types are not supported for the time being. The default subscription type is ``STREAM``.

```java
private Map<String, String> extendConfig = new HashMap<>()

```

   Indicates extended configuration, reserved field.

### public class Element

Element class that contains element id and data cache.

#### Private Members

```java
private long id

```

   The id of the element.

```java
private ByteBuffer buffer

```

   Data cache.

## Stream Object Operation Interface

### public static Producer createProducer(String streamName) throws YRException

Create a producer.

- Parameters:

   - **streamName** - The name of the stream. The length must be less than 256 characters and contain only the following characters ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``.

- Returns:

    Producer: Producer Interface.

- Throws:

   - **YRException** - Unified exception types thrown.

### public static Producer createProducer(String streamName, ProducerConfig producerConf) throws YRException

Create a producer.

- Parameters:

   - **streamName** - The name of the stream. The length must be less than 256 characters and contain only the following characters ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``.
   - **producerConf** - Producer configuration information.

- Returns:

    Producer: Producer Interface.

- Throws:

   - **YRException** - Unified exception types thrown.

### public static Consumer subscribe(String streamName, SubscriptionConfig config) throws YRException

Create a consumer.

- Parameters:

   - **streamName** - The name of the stream. The length must be less than 256 characters and contain only the following characters ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``.
   - **config** - Consumer configuration information.

- Returns:

    Consumer: Consumer interface.

- Throws:

   - **YRException** - Unified exception types thrown.

### public static Consumer subscribe(String streamName, SubscriptionConfig config, boolean autoAck) throws YRException

Create a consumer.

- Parameters:

   - **streamName** - The name of the stream. The length must be less than 256 characters and contain only the following characters ``(a-zA-Z0-9\\~\\.\\-\\/_!@#%\\^\\&\\*\\(\\)\\+\\=\\:;)``.
   - **config** - Consumer configuration information.
   - **autoAck** - When `autoAck` = ``true``, the consumer automatically sends an Ack to the data system for the previous message when it receives a message.

- Returns:

    Consumer: Consumer interface.

- Throws:

   - **YRException** - Unified exception types thrown.

## public interface Producer

Producer interface class.

### Interface description

#### void send(Element element) throws YRException

The producer sends data, which is first placed in a buffer.

The buffer is flushed according to the configured automatic flush policy (send at a certain interval or when the buffer is full) or by actively calling flush to allow consumers to access it.

- Parameters:

   - **element** - The Element data to be sent. Element can refer to the Element object structure in the public structure.

- Throws:

   - **YRException** - Unified exception types thrown.

#### void send(Element element, int timeoutMs) throws YRException

The producer sends data, which is first placed in a buffer.

The buffer is flushed according to the configured automatic flush policy (send at a certain interval or when the buffer is full) or by actively calling flush to allow consumers to access it.

- Parameters:

   - **element** - The Element data to be sent. Element can refer to the Element object structure in the public structure.
   - **timeoutMs** - Timeout period.

- Throws:

   - **YRException** - Unified exception types thrown.

#### void close() throws YRException

Closing a producer triggers an automatic flush of the data buffer and indicates that the data buffer is no longer in use.

Once closed, the producer can no longer be used.

- Throws:

   - **YRException** - Unified exception types thrown.

## public interface Consumer

Consumer interface class.

### Interface description

#### List<Element> receive(long expectNum, int timeoutMs) throws YRException

The consumer receives data with a subscription function.

The consumer waits for expectNum elements. The call returns when the timeout time timeoutMs is reached or the expected number of data is received.

- Parameters:

   - **expectNum** - The number of elements expected to be received.
   - **timeoutMs** - Timeout for receiving.

- Returns:

    List<Element>, A list of Elements that store data.

- Throws:

   - **YRException** - Unified exception types thrown.

#### List<Element> receive(int timeoutMs) throws YRException

The consumer receives data with a subscription function.

The call returns when the timeout time timeoutMs is reached.

- Parameters:

   - **timeoutMs** - Timeout for receiving.

- Returns:

    List<Element>, A list of Elements that store data.

- Throws:

   - **YRException** - Unified exception types thrown.

#### void ack(long elementId) throws YRException

After a consumer finishes using an element identified by a certain elementId, it needs to confirm that it has finished consuming, so that each worker can obtain information on whether all consumers have finished consuming.

If a certain page has been consumed, the internal memory recovery mechanism can be triggered. If not ack it will be automatically ack when the consumer exits.

- Parameters:

   - **elementId** - The id of the consumed element to be confirmed.

- Throws:

   - **YRException** - Unified exception types thrown.

#### void close() throws YRException

Close the consumer. Once closed, the consumer cannot be used.

- Throws:

   - **YRException** - Unified exception types thrown.

## Example

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
        // handle empty.
    }
    Element e = recv.get(0);
    Charset charset = Charset.forName("UTF-8");
    String res = charset.decode(e.getBuffer()).toString();
    consumer.ack(e.getId());

    producer.close();
    consumer.close();
    YR.deleteStream("aaaaaaaa");
} catch (YRException e) {
    // handle exception.
}
```
