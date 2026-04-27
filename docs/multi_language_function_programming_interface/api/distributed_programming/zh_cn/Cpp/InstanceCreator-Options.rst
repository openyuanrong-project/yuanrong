InstanceCreator::Options
=========================

.. cpp:function:: inline InstanceCreator<Creator>\
                  &YR::internal::InstanceCreator::Options(InvokeOptions &&optsInput)

    设置实例创建的选项，包括资源使用设置。
    
    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);
 
           YR::InvokeOptions opts;
           opts.cpu = 1000;
           opts.memory = 1000;
           auto ins = YR::Instance(SimpleCaculator::Constructor).Options(opts).Invoke();
           auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);
 
           return 0;
       }

    有状态函数指定本地执行（嵌套无状态）。

    .. code-block:: cpp

       // Example: Setting options for instance creation
       int incre(int x)
       {
            return x + 1;
       }
 
       YR_INVOKE(incre)
 
       struct Actor {
           int a;
 
           Actor() {}
           Actor(int init) { a = init; }
           static Actor *Create(int init) {
               return new Actor(init);
           }
           int InvokeIncre(bool local) {
               auto obj = YR::Function(incre).Options({ .alwaysLocalMode = local }).Invoke(a);
               return *YR::Get(obj);
           }
           MSGPACK_DEFINE(a)
       };
 
       YR_INVOKE(Actor::Create, &Actor::InvokeIncre)
 
       void Forward(bool local)
       {
           auto ins = YR::Instance(Actor::Create).Options({ .alwaysLocalMode = local }).Invoke(0);
           auto obj = ins.Function(&Actor::InvokeIncre).Invoke(true);
           std::cout << *YR::Get(obj) << std::endl;
           auto obj2 = ins.Function(&Actor::InvokeIncre).Invoke(false);
           std::cout << *YR::Get(obj2) << std::endl;
       }
 
       int main()
       {
           Forward(true);
           Forward(false);
           return 0;
       }

    .. warning::
        此方法专为分布式环境设计，在本地模式下不起作用。请确保在适用资源管理和执行选项的适当上下文中使用它。

    .. note::
        此方法在本地模式下无效。它专为分布式场景设计，在这些场景中，资源管理和执行选项至关重要。

    模板参数：
        - **Creator** - 用于构建实例的工作函数的类型。
  
    参数：
        - **optsInput** - 调用选项，包括资源使用和执行模式。详细描述请参见 :doc:`结构体说明 <./struct-Config>`。

    返回：
        InstanceCreator<Creator>&： ``InstanceCreator`` 对象的引用，便于直接调用 ``Invoke`` 方法。。