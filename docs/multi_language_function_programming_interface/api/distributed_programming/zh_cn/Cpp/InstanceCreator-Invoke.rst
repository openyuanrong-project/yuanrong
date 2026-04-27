InstanceCreator::Invoke
========================

.. cpp:function:: template<typename ...Args>\
                  inline NamedInstance<GetClassOfWorker<Creator>>\
                  YR::internal::InstanceCreator::Invoke(Args&&... args)

    执行实例的创建并构建类的对象。
    
    .. code-block:: cpp

       // 示例：创建一个实例并调用其成员函数
       int main(void) {
           YR::Config conf;
           YR::Init(conf);
   
           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke(); // 创建实例
           auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);    // 调用成员函数
           int res = *(YR::Get(r3));                                     // 获取结果
           return 0;
       }

    .. warning::
        确保提供的参数与构造函数预期的类型和数量完全匹配，以避免编译错误。通过此方法创建的实例仅设计用于与类的成员函数配合使用。

    .. note::
        此方法构建类的一个实例，并确保该实例只能执行属于该类的成员函数请求。返回的命名实例允许你调用该类的成员函数。

    模板参数：
        - **Creator** - 用于构建实例的工作函数的类型。
        - **Args** - 工作函数的构造函数所需参数的类型。
  
    参数：
        - **args** - 工作函数的构造函数所需的参数。参数的类型和数量需与构造函数的定义完全匹配。确保传递的参数类型精确符合预期，以避免因隐式类型转换导致的问题。

    返回：
        NamedInstance<GetClassOfWorker<Creator>>：一个命名实例，可以通过 `Function` 方法调用类的成员函数，适用于面向对象编程场景。