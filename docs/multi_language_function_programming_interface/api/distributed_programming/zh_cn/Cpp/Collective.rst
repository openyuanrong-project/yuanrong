Collective
==========

CollectiveGroup
---------------

类型和枚举定义
----------------

.. cpp:enum:: YR::Collective::Backend : uint8_t

    集合通信后端类型。

    .. cpp:enumerator:: GLOO = 0

        使用 GLOO 后端进行集合通信。

    .. cpp:enumerator:: INVALID

        无效的后端类型。

.. cpp:enum:: YR::DataType : uint8_t

    数据类型枚举，定义在 ``reduce_op.h`` 中。

    .. cpp:enumerator:: INT = 0

        整数类型（int）。

    .. cpp:enumerator:: DOUBLE

        双精度浮点数类型（double）。

    .. cpp:enumerator:: LONG

        长整数类型（long）。

    .. cpp:enumerator:: FLOAT

        单精度浮点数类型（float）。

    .. cpp:enumerator:: INVALID

        无效的数据类型。

.. cpp:enum:: YR::ReduceOp : uint8_t

    规约操作符枚举，定义在 ``reduce_op.h`` 中。

    .. cpp:enumerator:: SUM = 0

        求和操作。

    .. cpp:enumerator:: PRODUCT = 1

        乘积操作。

    .. cpp:enumerator:: MIN = 2

        最小值操作。

    .. cpp:enumerator:: MAX = 3

        最大值操作。

参数结构补充说明如下：

.. cpp:struct:: YR::Collective::CollectiveGroupSpec

    集合通信组的配置规格。

    **公共成员**

    .. cpp:member:: int worldSize

        组中的进程总数（world size）。

    .. cpp:member:: std::string groupName = "default"

        组的名称，默认为 "default"。组名必须匹配正则表达式：``^[a-zA-Z0-9\-_!#%\^\*\(\)\+\=\:;]*$``

    .. cpp:member:: Backend backend = Backend::GLOO

        使用的后端类型，默认为 GLOO。

    .. cpp:member:: int timeout = DEFAULT_COLLECTIVE_TIMEOUT

        操作超时时间（毫秒），默认为 60000 毫秒（60 秒）。

常量定义
---------

.. cpp:var:: const int YR::Collective::DEFAULT_COLLECTIVE_TIMEOUT

    默认的集合通信超时时间，值为 60000 毫秒（60 秒）。

.. cpp:var:: const std::string YR::Collective::DEFAULT_GROUP_NAME

    默认的组名称，值为 "default"。

Collective-GroupOps
-------------------

集合通信组管理接口，用于创建、初始化和销毁集合通信组。

InitCollectiveGroup
--------------------

.. cpp:function:: void YR::Collective::InitCollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, \
                                                           const std::string &prefix = "")

    在 actor 实例中初始化集合通信组。

    此函数用于在当前 actor 实例中初始化一个集合通信组。通常在分布式训练或并行计算场景中，每个进程可以调用此函数来加入集合通信组。

    .. code-block:: cpp

       YR::Collective::CollectiveGroupSpec spec;
       spec.worldSize = 4;
       spec.groupName = "my_group";
       spec.backend = YR::Collective::Backend::GLOO;
       spec.timeout = 60000;

       int rank = 0;  // 当前进程的 rank
       YR::Collective::InitCollectiveGroup(spec, rank);

       // 使用集合通信操作...
       return 0;


    参数：
        - **groupSpec** - 集合通信组的配置规格，包含 worldSize、groupName、backend 和 timeout。
        - **rank** - 当前进程在组中的排名（rank），范围应为 [0, worldSize-1]。
        - **prefix** - 存储前缀，用于后端通信的键值存储。默认为空字符串。

    .. note::
        限制条件

        - 必须在调用 ``YR::Init()`` 之后才能调用此函数。
        - 同一个 groupName 不能重复初始化，否则会抛出异常。
        - groupName 必须匹配正则表达式：``^[a-zA-Z0-9\-_!#%\^\*\(\)\+\=\:;]*$``
        - 不支持 driver 中的 ``CreateCollectiveGroup`` 和 actor 中的 ``InitCollectiveGroup`` 混合加入同一个 group。同一个 group 必须全部使用 ``CreateCollectiveGroup`` 创建，或者全部使用 ``InitCollectiveGroup`` 初始化。
        - 不支持动态增删 group 中的成员。group 创建后，成员数量固定，无法在运行时添加或删除成员。

    抛出：
        :cpp:class:`Exception` - 以下情形会抛出异常：

        - 如果在初始化前调用此函数，将抛出异常。
        - 如果 groupName 无效，将抛出异常。
        - 如果集合通信组已存在，将抛出异常。

CreateCollectiveGroup
----------------------

.. cpp:function:: void YR::Collective::CreateCollectiveGroup(const CollectiveGroupSpec &groupSpec, \
                                                              const std::vector<std::string> &instanceIDs, \
                                                              const std::vector<int> &ranks)

    在 driver 中创建集合通信组，使用 actor 实例 ID 列表。

    此函数用于在 driver 进程中创建集合通信组，指定参与集合通信的 actor 实例及其对应的 rank。通常在分布式训练场景中，driver 进程会调用此函数来创建组，然后各个 actor 实例不用通过 InitCollectiveGroup 加入组。

    .. code-block:: cpp

       #include "yr/collective/collective.h"

       int main() {
           YR::Config conf;
           YR::Init(conf);

           // 创建包含4个actor的集合通信组
           std::vector<YR::NamedInstance<CollectiveActor>> instances;
           std::vector<std::string> instanceIDs;
           for (int i = 0; i < 4; ++i) {
               auto ins = YR::Instance(CollectiveActor::FactoryCreate).Invoke();
               instances.push_back(ins);
               instanceIDs.push_back(ins.GetInstanceId());
           }

           std::string groupName = "test-group";
           YR::Collective::CollectiveGroupSpec spec{
               .worldSize = 4,
               .groupName = groupName,
               .backend = YR::Collective::Backend::GLOO,
               .timeout = YR::Collective::DEFAULT_COLLECTIVE_TIMEOUT,
           };
           YR::Collective::CreateCollectiveGroup(spec, instanceIDs, {0, 1, 2, 3});

           // 调用每个actor上的计算方法
           std::vector<int> input = {1, 2, 3, 4};
           std::vector<YR::ObjectRef<int>> res;
           for (int i = 0; i < 4; ++i) {
               res.push_back(instances[i]
                                 .Function(&CollectiveActor::Compute)
                                 .Invoke(input, groupName, static_cast<uint8_t>(YR::ReduceOp::SUM)));
           }

           // Get 结果
           auto res0 = *YR::Get(res[0]);
           auto res1 = *YR::Get(res[1]);
           std::cout << "AllReduce result: " << res0 << ", Recv result: " << res1 << std::endl;

           // 销毁组
           YR::Collective::DestroyCollectiveGroup(groupName);
           YR::Finalize();
           return 0;


    参数：
        - **groupSpec** - 集合通信组的配置规格。
        - **instanceIDs** - actor 实例 ID 列表，大小必须等于 worldSize。
        - **ranks** - 每个实例对应的 rank 列表，大小必须等于 worldSize，且 ranks 的值应在 [0, worldSize-1] 范围内。

    .. note::
        限制条件

        - instanceIDs 的大小必须等于 worldSize。
        - ranks 的大小必须等于 worldSize。
        - 如果 groupName 已存在，将抛出异常，需要先调用 DestroyCollectiveGroup 销毁已存在的组。
        - 不支持 driver 中的 ``CreateCollectiveGroup`` 和 actor 中的 ``InitCollectiveGroup`` 混合加入同一个 group。同一个 group 必须全部使用 ``CreateCollectiveGroup`` 创建，或者全部使用 ``InitCollectiveGroup`` 初始化。
        - 不支持动态增删 group 中的成员。group 创建后，成员数量固定，无法在运行时添加或删除成员。

    抛出：
        :cpp:class:`Exception` - 以下情形会抛出异常：

        - 如果 instanceIDs、ranks 和 worldSize 不匹配，将抛出异常。
        - 如果 groupName 已存在，将抛出异常。
        - 如果 groupName 无效，将抛出异常。

DestroyCollectiveGroup
-----------------------

.. cpp:function:: void YR::Collective::DestroyCollectiveGroup(const std::string &groupName)

    销毁指定的集合通信组。

    此函数用于销毁一个已创建的集合通信组，释放相关资源。

    参数：
        - **groupName** - 要销毁的组的名称。

    .. note::
        限制条件

        - 如果组不存在，此函数不会抛出异常，会静默处理。

GetWorldSize
------------

.. cpp:function:: int YR::Collective::GetWorldSize(const std::string &groupName = DEFAULT_GROUP_NAME)

    获取指定组中的进程总数。

    此函数返回集合通信组中的进程总数（world size）。

    .. code-block:: cpp

       // 初始化集合通信组...
       std::string groupName = "my_group";
       int worldSize = YR::Collective::GetWorldSize(groupName);
       std::cout << "World size: " << worldSize << std::endl;

    参数：
        - **groupName** - 组的名称，默认为 "default"。

    返回：
        组中的进程总数（world size）。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

GetRank
-------

.. cpp:function:: int YR::Collective::GetRank(const std::string &groupName = DEFAULT_GROUP_NAME)

    获取当前进程在指定组中的排名。

    此函数返回当前进程在集合通信组中的排名（rank）。

    .. code-block:: cpp

       // 初始化集合通信组...
       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       std::cout << "My rank: " << rank << std::endl;

    参数：
        - **groupName** - 组的名称，默认为 "default"。

    返回：
        当前进程在组中的排名（rank），范围在 [0, worldSize-1]。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。
        
Collective-CommOps
------------------

集合通信操作接口，提供 AllReduce、Reduce、AllGather、Broadcast 等分布式通信原语。

AllReduce
----------

.. cpp:function:: void YR::Collective::AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, \
                                                 const ReduceOp &op, const std::string &groupName = DEFAULT_GROUP_NAME)

    在所有进程中执行规约操作，并将结果广播到所有进程。

    此函数在所有进程中执行规约操作（如求和、求最大值等），并将结果写回到所有进程的 recvbuf 中。所有进程的 recvbuf 最终包含相同的结果。

    .. code-block:: cpp

       int Compute(std::vector<int> in, std::string &groupName, uint8_t op)
       {
       std::string groupName = "my_group";
       std::vector<int> localData(100, 1);
       std::vector<int> result(100);

       // AllReduce：在所有进程执行规约并广播结果
       YR::Collective::AllReduce(localData.data(), result.data(), localData.size(), YR::DataType::INT, YR::ReduceOp::SUM, groupName);

       // Barrier：同步所有进程
       YR::Collective::Barrier(groupName);
       YR::Collective::DestroyCollectiveGroup(groupName);
       int res = 0;
       for (int i = 0; i < localData.size(); ++i) {
           res += result[i];
       }
       return res;
       }

    参数：
        - **sendbuf** - 发送缓冲区，包含本地输入数据。所有进程的 sendbuf 大小应相同。
        - **recvbuf** - 接收缓冲区，用于存储规约结果。大小应与 sendbuf 相同。
        - **count** - 数据元素的数量。
        - **dtype** - 数据类型（DataType::INT、DataType::FLOAT、DataType::DOUBLE、DataType::LONG）。
        - **op** - 规约操作符（ReduceOp::SUM、ReduceOp::PRODUCT、ReduceOp::MIN、ReduceOp::MAX）。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - sendbuf 和 recvbuf 可以指向同一块内存（in-place 操作）。
        - 所有进程必须调用此函数，且参数 count、dtype、op 必须一致。
        - 必须在组创建并初始化后才能调用此函数。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Reduce
------

.. cpp:function:: void YR::Collective::Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, \
                                              const ReduceOp &op, int dstRank, \
                                              const std::string &groupName = DEFAULT_GROUP_NAME)

    在所有进程中执行规约操作，并将结果发送到指定的目标进程。

    此函数在所有进程中执行规约操作，但只有 dstRank 进程的 recvbuf 会包含规约结果，其他进程的 recvbuf 内容未定义。

    .. code-block:: cpp

       std::string groupName = "my_group";
       std::vector<int> localData(100, 1);
       std::vector<int> result(100);

       int rootRank = 0;  // 规约结果发送到 rank 0
       YR::Collective::Reduce(localData.data(), result.data(), localData.size(), YR::DataType::INT, YR::ReduceOp::SUM,
                              rootRank, groupName);

       int rank = YR::Collective::GetRank(groupName);
       if (rank == rootRank) {
           // 只有 rootRank 的 result 包含有效结果
           std::cout << "Reduced result: " << result[0] << std::endl;
       }

    参数：
        - **sendbuf** - 发送缓冲区，包含本地输入数据。
        - **recvbuf** - 接收缓冲区，仅在 dstRank 进程中使用，用于存储规约结果。其他进程的 recvbuf 内容未定义。
        - **count** - 数据元素的数量。
        - **dtype** - 数据类型。
        - **op** - 规约操作符。
        - **dstRank** - 目标进程的 rank，规约结果将发送到此进程。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - 非 root rank（非 dstRank）的 recvbuf 输出数据不可信，请不要使用。只有 dstRank 进程的 recvbuf 包含有效的规约结果。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

AllGather
---------

.. cpp:function:: void YR::Collective::AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype, \
                                                  const std::string &groupName = DEFAULT_GROUP_NAME)

    从所有进程收集数据，并将结果广播到所有进程。

    此函数从所有进程收集数据，并将收集到的数据按 rank 顺序排列后写回到所有进程的 recvbuf 中。recvbuf 的大小应为 count * worldSize。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       int worldSize = YR::Collective::GetWorldSize(groupName);

       // 每个进程发送不同的数据
       std::vector<float> localData(10, static_cast<float>(rank));
       std::vector<float> gatheredData(10 * worldSize);

       YR::Collective::AllGather(localData.data(), gatheredData.data(), localData.size(), YR::DataType::FLOAT,
                                 groupName);

    参数：
        - **sendbuf** - 发送缓冲区，包含本地要发送的数据。
        - **recvbuf** - 接收缓冲区，用于存储从所有进程收集的数据。大小应为 count * worldSize。
        - **count** - 每个进程发送的数据元素数量。
        - **dtype** - 数据类型。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - recvbuf 的大小必须至少为 count * worldSize。
        - 收集的数据按 rank 顺序排列在 recvbuf 中。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Broadcast
---------

.. cpp:function:: void YR::Collective::Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, \
                                                  int srcRank, const std::string &groupName = DEFAULT_GROUP_NAME)

    从源进程广播数据到所有进程。

    此函数从 srcRank 进程广播数据到组中的所有进程。所有进程的 recvbuf 最终包含相同的数据（来自 srcRank 的 sendbuf）。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       int srcRank = 0;

       std::vector<int> data(100);
       if (rank == srcRank) {
           // 源进程初始化数据
           for (int i = 0; i < 100; i++) {
               data[i] = i;
           }
       }

       YR::Collective::Broadcast(data.data(), data.data(), data.size(), YR::DataType::INT, srcRank, groupName);
       // All processes' data now contains the same content (from srcRank)

    参数：
        - **sendbuf** - 发送缓冲区，仅在 srcRank 进程中使用，包含要广播的数据。
        - **recvbuf** - 接收缓冲区，用于存储广播的数据。所有进程的 recvbuf 最终包含相同的数据。
        - **count** - 数据元素的数量。
        - **dtype** - 数据类型。
        - **srcRank** - 源进程的 rank。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - sendbuf 和 recvbuf 可以指向同一块内存（in-place 操作）。
        - 所有进程必须调用此函数，且参数 count、dtype、srcRank 必须一致。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Scatter
-------

.. cpp:function:: void YR::Collective::Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, \
                                               DataType dtype, int srcRank, \
                                               const std::string &groupName = DEFAULT_GROUP_NAME)

    从源进程分散数据到所有进程。

    此函数从 srcRank 进程的 sendbuf 向量中分散数据到所有进程。srcRank 进程的 sendbuf 向量应包含 worldSize 个缓冲区，每个缓冲区对应一个目标 rank。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       int worldSize = YR::Collective::GetWorldSize(groupName);
       int srcRank = 0;

       std::vector<int> recvData(10);

       if (rank == srcRank) {
           // 源进程准备要分发的数据
           std::vector<std::vector<int>> sendData(worldSize, std::vector<int>(10));
           for (int i = 0; i < worldSize; i++) {
               for (int j = 0; j < 10; j++) {
                   sendData[i][j] = i * 10 + j;
               }
           }

           std::vector<void *> sendbuf;
           for (auto &vec : sendData) {
               sendbuf.push_back(vec.data());
           }

           YR::Collective::Scatter(sendbuf, recvData.data(), 10, YR::DataType::INT, srcRank, groupName);
       } else {
           // 非源进程传入空向量
           std::vector<void *> sendbuf;  // Can be empty
           YR::Collective::Scatter(sendbuf, recvData.data(), 10, YR::DataType::INT, srcRank, groupName);
       }

       // 每个进程的 recvData 包含来自 srcRank 的数据

    参数：
        - **sendbuf** - 发送缓冲区向量，仅在 srcRank 进程中使用。大小应为 worldSize，每个元素指向要发送给对应 rank 的数据。
        - **recvbuf** - 接收缓冲区，用于存储从 srcRank 接收的数据。
        - **count** - 每个进程接收的数据元素数量。
        - **dtype** - 数据类型。
        - **srcRank** - 源进程的 rank。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - sendbuf 向量仅在 srcRank 进程中使用，其他进程可以传入空向量。
        - sendbuf 向量的大小必须等于 worldSize。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Barrier
-------

.. cpp:function:: void YR::Collective::Barrier(const std::string &groupName = DEFAULT_GROUP_NAME)

    同步屏障，阻塞直到组中的所有进程都到达此点。

    此函数用于同步所有进程，确保所有进程在执行后续代码之前都到达了 Barrier 调用点。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);

       // 执行一些异步操作...
       std::cout << "Rank " << rank << " before barrier" << std::endl;

       YR::Collective::Barrier(groupName);

       // 所有进程在此同步等待
       std::cout << "Rank " << rank << " after barrier" << std::endl;

    参数：
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - 所有进程必须调用此函数，否则会导致死锁。
        - 此函数会阻塞，直到所有进程都到达 Barrier 调用点。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Send
----

.. cpp:function:: void YR::Collective::Send(const void *sendbuf, int count, DataType dtype, int dstRank, \
                                           int tag = 0, const std::string &groupName = DEFAULT_GROUP_NAME)

    向指定进程发送数据。

    此函数用于点对点通信，向指定的目标进程发送数据。需要与 Recv 函数配对使用，通过 tag 参数匹配对应的接收操作。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       int worldSize = YR::Collective::GetWorldSize(groupName);

       if (rank == 0) {
           // Rank 0 向 Rank 1 发送数据
           std::vector<float> data(100, 1.0f);
           YR::Collective::Send(data.data(), data.size(), YR::DataType::FLOAT, 1, 0, groupName);
       } else if (rank == 1) {
           // Rank 1 接收 Rank 0 发送的数据
           std::vector<float> recvData(100);
           YR::Collective::Recv(recvData.data(), recvData.size(), YR::DataType::FLOAT, 0, 0, groupName);
       }

    参数：
        - **sendbuf** - 发送缓冲区，包含要发送的数据。
        - **count** - 要发送的数据元素数量。
        - **dtype** - 数据类型。
        - **dstRank** - 目标进程的 rank。
        - **tag** - 消息标签，用于匹配对应的 Recv 操作，默认为 0。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - Send 和 Recv 必须配对使用，且 tag 必须匹配。
        - Send 操作会阻塞，直到对应的 Recv 操作被调用。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

Recv
----

.. cpp:function:: void YR::Collective::Recv(void *recvbuf, int count, DataType dtype, int srcRank, \
                                           int tag = 0, const std::string &groupName = DEFAULT_GROUP_NAME)

    从指定进程接收数据。

    此函数用于点对点通信，从指定的源进程接收数据。需要与 Send 函数配对使用，通过 tag 参数匹配对应的发送操作。

    .. code-block:: cpp

       std::string groupName = "my_group";
       int rank = YR::Collective::GetRank(groupName);
       int worldSize = YR::Collective::GetWorldSize(groupName);

       if (rank == 0) {
           // Rank 0 sends data to Rank 1
           std::vector<float> data(100, 1.0f);
           YR::Collective::Send(data.data(), data.size(), YR::DataType::FLOAT, 1, 0, groupName);
       } else if (rank == 1) {
           // Rank 1 receives data from Rank 0
           std::vector<float> recvData(100);
           YR::Collective::Recv(recvData.data(), recvData.size(), YR::DataType::FLOAT, 0, 0, groupName);
       }

    参数：
        - **recvbuf** - 接收缓冲区，用于存储接收到的数据。
        - **count** - 要接收的数据元素数量。
        - **dtype** - 数据类型。
        - **srcRank** - 源进程的 rank。
        - **tag** - 消息标签，用于匹配对应的 Send 操作，默认为 0。
        - **groupName** - 组的名称，默认为 "default"。

    .. note::
        注意事项

        - Recv 和 Send 必须配对使用，且 tag 必须匹配。
        - Recv 操作会阻塞，直到对应的 Send 操作被调用。
        - count 和 dtype 必须与对应的 Send 操作一致。

    抛出：
        :cpp:class:`Exception` - 如果组不存在或尚未创建，将抛出异常。

