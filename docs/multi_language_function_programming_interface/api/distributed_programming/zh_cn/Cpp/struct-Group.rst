Group
=======

.. cpp:class:: Group

    一个用于管理成组实例生命周期的类。

    Group 类负责管理成组实例的生命周期，包括它们的创建和销毁。它遵循命运共享原则，即组中的所有实例将一起创建或销毁。

    Group 类提供了用于创建、终止和管理成组实例的方法。它确保组中的所有实例被视为一个整体，并且在组创建过程中如果出现任何失败，整个组的操作将回滚。

    **公共函数**
 
    .. cpp:function:: Group() = default
       
        默认构造函数。

    .. cpp:function:: Group(std::string &name, GroupOptions &opts)
       
        构造函数。

        参数：
            - **name** - 实例组调度类名称，名称必须是唯一的。
            - **opts** - 参见 GroupOptions 结构体。
    
    .. cpp:function:: Group(std::string &name, GroupOptions &&opts)
       
        构造函数。
    
        参数：
            - **name** - 实例组调度类名称，名称必须是唯一的。
            - **opts** - 参见 GroupOptions 结构体。

    .. cpp:function:: void Invoke()
       
        按照命运共享原则执行实例组的创建。

        此函数用于创建一组共享相同命运的实例。组中的所有实例将一起创建，如果其中一个实例创建失败，整个组将回滚。组的配置和选项定义在结构体 GroupOptions 中。
        
        .. code-block:: cpp

           int main(void) {
           YR::Config conf;
           YR::Init(conf);
           std::string groupName = "";
           YR::InvokeOptions opts;
           opts.groupName = groupName;
           YR::GroupOptions groupOpts;
           groupOpts.timeout = 60;
           groupOpts.strategy = "SPREAD"
           groupOpts.bind.resource = "NUMA";
           groupOpts.bind.strategy = "PACK";
           auto g = YR::Group(groupName, groupOpts);
           auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
           auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
           g.Invoke();
           return 0;
         }

        .. note::
            限制条件

            - 单个组最多可以创建 256 个实例。
            - 并发创建支持最多 12 个组，每个组最多创建 256 个实例。
            - 在调用 `NamedInstance::Export() <NamedInstance.html#classYR_1_1NamedInstance_1a9312f83438dbe23880e0dbddf85f0c39>`_ 之后调用此接口，会导致当前线程挂起。
            - 如果未调用此接口，直接调用有状态函数请求并获取结果，会导致当前线程挂起。
            - 重复调用此接口将导致抛出异常。
            - 组中的实例不支持指定独立生命周期。

        抛出：
            **std::runtime_error** - 如果在组创建过程中发生错误，或者接口被误用，将抛出此异常。
  
        返回：
            void。

    .. cpp:function:: void Terminate()
       
        终止一组实例。

        此函数用于删除或终止一组一起创建的实例。根据命运共享原则，组中的所有实例将作为一个整体被清理或移除。
    
            
        .. code-block:: cpp
    
           int main(void) {
             YR::Config conf;
             YR::Init(conf);
             std::string groupName = "";
             YR::InvokeOptions opts;
             opts.groupName = groupName;
             YR::GroupOptions groupOpts;
             auto g = YR::Group(groupName, groupOpts);
             auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
             auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
             g.Invoke();
             g.Terminate();
             return 0;
           }
    
        .. note::
            限制条件
    
            - 此函数只能在已成功创建并调用的组上调用。
            - 并发创建支持最多 12 个组，每个组最多创建 256 个实例。
            - 在调用 `NamedInstance::Export() <NamedInstance.html#classYR_1_1NamedInstance_1a9312f83438dbe23880e0dbddf85f0c39>`_ 之后调用此接口，会导致当前线程挂起。
            - 如果未调用此接口，直接调用有状态函数请求并获取结果，会导致当前线程挂起。
            - 重复调用此接口将导致抛出异常。
            - 组中的实例不支持指定独立生命周期。
    
        抛出：
            **std::runtime_error** - 如果组不存在、已经被终止，或者在终止过程中发生错误，将抛出异常。
      
        返回：
            void。
    
    .. cpp:function:: void Wait()
       
        等待成组实例的创建和执行完成。

        此函数会阻塞当前进程，直到组中的所有实例完成它们的创建和执行。如果在调用 `Wait` 之前没有调用 `Invoke` 方法，将抛出异常。
            
        .. code-block:: cpp
            
           int main(void) {
               YR::Config conf;
               YR::Init(conf);
               std::string groupName = "";
               YR::InvokeOptions opts;
               opts.groupName = groupName;
               YR::GroupOptions groupOpts;
               auto g = YR::Group(groupName, groupOpts);
               auto ins1 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
               auto ins2 = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
               g.Invoke();
               g.Wait();  // Wait for all instances in the group to complete
               return 0;
           }
            
        抛出：
            :cpp:class:`Exception` - 如果在调用 `Wait` 之前未调用 `Invoke` 方法，或者等待操作超时，则会抛出异常。超时时间在 `GroupOptions` 结构体中指定。
              
        返回：
            void。

    .. cpp:function:: std::string GetGroupName() const

        获取组的名称。

        返回：
            组的名称。

参数结构补充说明如下：

.. cpp:struct:: GroupOptions
    
    成组实例调度的配置选项。

    GroupOptions 结构体定义了成组实例生命周期管理的参数，包括当内核资源不足时重新调度的超时设置。

    **公共成员**
 
    .. cpp:member:: int timeout = NO_TIMEOUT
       
       当内核资源不足时重新调度的超时时间，单位为秒。

       - 如果设置为 **-1**，内核将无限期地重试调度。
       - 如果设置为小于 **0** 的值（不包括 -1），将抛出异常。
  
    .. cpp:member:: bool sameLifecycle = true
       
       是否为成组实例启用命运共享配置。

       - **true**：组中的实例将一起创建和销毁，此为默认值。
       - **false**：实例可以拥有独立的生命周期。

    .. cpp:member:: std::string strategy = "None"

       组创建的策略。

       - **None**：无策略。
       - **PACK**：尽可能将多个实例打包到同一个节点中。
       - **SPREAD**：尽可能将多个实例分布在不同的节点上。
       - **STRICT_PACK**：所有实例必须放置在同一个节点上，否则创建将失败。
       - **STRICT_SPREAD**：所有实例必须放置在不同的节点上，否则创建将失败。

    .. cpp:member:: BindOptions bind

       资源绑定选项。用于指定资源级别的绑定策略。

.. cpp:struct:: BindOptions

    资源绑定亲和性配置。

    .. note::
        限制条件

        - 资源亲和功能必须与组调度一起使用，不支持在单个实例的调度选项中配置资源亲和性。
        - 在非 NUMA 系统上配置 NUMA 绑定会导致绑定失败。
        - 部署参数 numa_collection_enable 需要设置为 true。

    **公共成员**

    .. cpp:member:: std::string resource = "NONE"

       要绑定的资源类型。默认类型为 ``None``。

       - **None**：不绑定。
       - **NUMA**：绑定到 NUMA 节点。

    .. cpp:member:: std::string strategy = "NONE"

       指定资源的绑定策略，默认策略为 ``None``。

       - **None**：等同于 ``SPREAD`` 。
       - **PACK**：尽可能将多个实例打包到同一个资源单元中。
       - **SPREAD**：尽可能将多个实例分布在不同的资源单元上。