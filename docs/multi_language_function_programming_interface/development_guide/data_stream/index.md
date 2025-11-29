# 数据流

openYuanrong提供了基于 Publish-SubScribe(pub-sub) 模型的数据流。通过数据流，应用可实现函数间无界数据流的数据交换，支持复杂的数据交互关系。同时，通过解耦数据生产方和消费方的同步连接关系，支持数据生产方和消费方按需异步调度。

数据流中有四个关键概念：生产者（producer）、消费者（consumer）、流（stream）、数据项（element）。

- 生产者：生产者是无界数据流的发起端，产生并发送数据。
- 消费者：消费者是无界数据流的接收端，消费数据。
- 数据项：数据以数据项的粒度在生产者和消费者间发送和接收。
- 流：生产者和消费者间不相互感知，通过流进行关联。生产者向流中发送数据，消费者订阅流并从流中接收数据。一个应用中可以有多条流，通过流名称区分。

数据流支持多生产者和多消费者间的数据交互。在实际场景中，使用最多的是 one-to-one（一个生产者一个消费者）、many-to-one（多个生产者一个消费者）及 one-to-many（一个生产者多个消费者）。

数据流接口支持 C++/Java/Python 多语言，本节以 C++ 语言为例介绍流、生产者、消费者、数据项四类接口的使用。

## 使用限制

- 数据流定位在支持计算过程的中间数据流转，当前无持久化能力。若发送节点故障，则数据会丢失，此时需应用层进行故障处理，比如重启业务等。
- 当前接口都是同步阻塞，没有类似 epoll 的多路复用能力。
- 由于通过流解耦了生产者和消费者，生产者和消费者间不感知对方。也就是当生产者关闭时，消费者是无法感知的，需要业务层进行处理。

## 流接口

流代表了生产者和消费者间的发布订阅交互关系。流随着生产者或消费者的创建而隐式创建，无需应用显式创建流。

创建一个生产者或消费者时，需要指定它所关联的流，不同的流通过流名称区分。如果流已经存在，新创建的生产者或消费者会关联到这条流上。如流不存在，在创建生产者或消费者时，系统会隐式创建一个新流，并关联它到指定的流名称。

一条流上的生产者和消费者都关闭后，如果未操作接口删除流，流仍然存在，应用可以在这条流上继续关联新的生产者和消费者。您可以在创建生产者时指定流为自动删除 (配置 `autoCleanup` 选项)，当该流上所有生产者和消费者都关闭后，流将被系统自动删除。

和流相关的操作有三个，创建生产者（可能隐式创建流）、创建消费者（可能隐式创建流）及删除流。

- 通过创建生产者产生流：`streamName` 为生产者关联的流名称，如果对应的流不存在，会隐式创建一个新的流。可通过 `ProducerConf` 配置生产者属性或采用默认配置。

    ```cpp
    // 接口原型
    std::shared_ptr<Producer> CreateProducer(const std::string &streamName, ProducerConf producerConf = {});
    ```

    ```cpp
    // 调用示例
    try {
        YR::ProducerConf producerConf{.delayFlushTime = 5, .pageSize = 1024 * 1024ul, .maxStreamSize = 1024 * 1024 * 1024ul};
        std::shared_ptr<YR::Producer> producer = YR::CreateProducer("streamName", producerConf);
    } catch (YR::Exception &e) {
        // .......
    }
    ```

- 通过创建消费者产生流：`streamName` 为消费者关联的流名称，如果对应的流不存在，会隐式创建一个新的流。`SubscriptionConfig` 用于配置消费者行为，`autoAck` 配置开启[自动 ACK](development-data-stream-ack)。

    ```cpp
    // 接口原型
    std::shared_ptr<Consumer> Subscribe(const std::string &streamName, const SubscriptionConfig &config, bool autoAck = false);
    ```

    ```cpp
    // 调用示例
    try {
        YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
        std::shared_ptr<YR::Consumer> consumer = YR::Subscribe("streamName", config, true);
    } catch (YR::Exception &e) {
        // .......
    }
    ```

- 删除流：在全局生产者和消费者个数为 0 时，该数据流不再使用，应用可使用流名称 `streamName` 主动删除流。

    ```cpp
    // 接口原型
    void DeleteStream(const std::string &streamName);
    ```

    ```cpp
    // 调用示例
    try {
        YR::DeleteStream("streamName");
    } catch (YR::Exception &e) {
        // .......
    }
    ```

## 生产者接口

生产者（Producer）可向流中发送数据。

- Send 方法：生产者发送数据，数据会先放入缓冲区，系统根据生产者配置的 Flush 策略 (发送间隔一段时间或者缓冲写满) 刷新缓冲使其对消费者可见。

- Close 方法：关闭生产者，会触发自动 Flush 数据缓冲。一旦关闭，生产者不再可用。

    ```cpp
    // 接口原型
    void Close();
    ```

    ```cpp
    // 调用示例
    try {
        std::shared_ptr<YR::Producer> producer = YR::CreateProducer("streamName");
        // .......
        producer->Close();
    } catch (YR::Exception &e) {
        // .......
    }
    ```

## 消费者接口

消费者（Consumer）可接收流中的数据，使用 `Ack` 方法确认数据接收，以及主动关闭。

- Receive 方法：接收数据会等待 `expectNum` 个 `elements`，当到达超时时间 `timeoutMs` 或满足期望个数的数据时，该调用返回。

    ```cpp
    // 接口原型
    void Receive(uint32_t timeoutMs, std::vector<Element> &outElements);
    void Receive(uint32_t expectNum, uint32_t timeoutMs, std::vector<Element> &outElements);

    ```

    ```cpp
    // 调用示例
    try {
        YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
        std::shared_ptr<YR::Consumer> consumer = YR::Subscribe("streamName", config);
        // ......
        std::vector<YR::Element> elements;
        consumer->Receive(1, 6000, elements);
    } catch (YR::Exception &e) {
        // .......
    }
    ```

(development-data-stream-ack)=

- Ack 方法：消费者接收数据后，需要对数据进行 ACK 操作，以确认该数据及之前收到的数据都已消费完。对确认消费完的数据，系统会回收内存资源。

    数据流提供了自动 ACK 功能，只需在创建消费者时配置 `autoAck` 为 `true`。应用每次调用 `Receive` 操作后，系统会自动确认上一次接收的数据，应用无需再主动调用 `Ack` 方法。

    :::{note}

    开启自动 ACK 后，用户需保证消费者每次调用 `Receive` 前，上一次接收到的数据已消费完。调用 `Receive` 后，继续消费上一次接收到的数据系统未定义。

    :::

    :::{hint}

    当消费者调用 `Receive` 方法时，会获取 `Element` 对象，对象的内部指针指向实际的数据，这些数据处于应用函数和数据系统间的共享内存之中。应用需要通过 ACK 操作确认该数据已经消费完，此时数据系统才能回收该数据所占的内存资源。如应用不调用 ACK 操作，数据系统无法判断数据是否被消费，则无法回收内存资源，最终导致内存资源耗尽，系统异常。

    :::

    ```cpp
    // 接口原型
    void Ack(uint64_t elementId);
    ```

    ```cpp
    // 调用示例
    try {
        YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
        std::shared_ptr<YR::Consumer> consumer = YR::Subscribe("streamName", config);
        // ......
        consumer->Ack("elementID");
    } catch (YR::Exception &e) {
        // .......
    }

    ```

- Close 方法：关闭消费者。一旦关闭，消费者不再可用。

    ```cpp
    // 接口原型
    void Close();
    ```

    ```cpp
    // 调用示例
    try {
        YR::SubscriptionConfig config("subName", YR::SubscriptionType::STREAM);
        std::shared_ptr<YR::Consumer> consumer = YR::Subscribe("streamName", config);
        // .......
        consumer->Close();
    } catch (YR::Exception &e) {
        // .......
    }
    ```

## 数据项

数据流以数据项为单位发送数据，数据项结构体如下:

| **字段** | **类型**  | **说明**       |
| -------- | --------- | -------------- |
| **ptr**  | uint8_t * | 指向数据的指针。 |
| **size** | unit64_t  | 数据长度。       |
| **id**   | uint64_t  | 数据项的 id。 |

`ptr` 字段指向数据的存放地址。在生产者端, `ptr` 字段指向应用的内存空间，因为发送的数据是应用填充的。在消费者端, `ptr` 字段指向的地址空间位于应用函数和数据系统间的共享内存中，这也是为什么会需要 ACK 操作的原因。
