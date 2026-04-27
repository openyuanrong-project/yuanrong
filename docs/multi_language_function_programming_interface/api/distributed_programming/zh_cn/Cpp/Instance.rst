Instance
=========

.. cpp:function:: template<typename F>\
                  std::enable_if_t<!std::is_base_of<internal::XLangClass, F>::value,\
                  YR::internal::InstanceCreator<F>> YR::Instance(F constructor)
    
    为构建一个类的实例创建一个 `InstanceCreator`。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke();
           auto r1 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);
           int res = *(YR::Get(r1));
           return 0;
        }

    模板参数：
        - **F** - 不能是继承自 `internal::XLangClass` 的类型。

    参数：
        - **constructor** - 一个用于构建类的实例并返回指向所构建对象的指针的函数。
    
    返回：
        InstanceCreator<F>：一个创建器对象，可用于创建实例并指定实例资源的额外选项。

.. cpp:function:: template<typename F>\
                  std::enable_if_t<!std::is_base_of<internal::XLangClass, F>::value,\
                  YR::internal::InstanceCreator<F>> YR::Instance(F constructor, const std::string &name)

    为构建一个类的实例创建一个 `InstanceCreator`。

    .. code-block:: cpp

        int main(void) {
           YR::Config conf;
           YR::Init(conf);

           // 具名实例的名字是 name_1
           auto counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(1);
           auto c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;

           // 生成 name_1 具名实例的句柄, 可以重用 name_1 具名实例
           counter = YR::Instance(Counter::FactoryCreate, "name_1").Invoke(1);
           c = counter.Function(&Counter::Add).Invoke(1);
           std::cout << "counter is " << *YR::Get(c) << std::endl;

           return 0;
        }

    模板参数：
        - **F** - 不能是继承自 `internal::XLangClass` 的类型。

    参数：
        - **constructor** - 一个用于构建类的实例并返回指向所构建对象的指针的函数。
        - **name** - 具名实例的实例名称，可以通过实例名称重用这个实例。

    返回：
        InstanceCreator<F>：一个创建器对象，可用于创建实例并指定实例资源的额外选项。

.. cpp:function:: template<typename F>\
                  std::enable_if_t<!std::is_base_of<internal::XLangClass, F>::value,\
                  YR::internal::InstanceCreator<F>> YR::Instance(F constructor, const std::string &name,\
                  const std::string &ns)

    为构建一个类的实例创建一个 `InstanceCreator`。

    .. code-block:: cpp

        int main(void) {
            YR::Config conf;
            YR::Init(conf);

            // 具名实例的名字是 ns_1-name_1
            auto counter = YR::Instance(Counter::FactoryCreate, "name_1", "ns_1").Invoke(1);
            auto c = counter.Function(&Counter::Add).Invoke(1);
            std::cout << "counter is " << *YR::Get(c) << std::endl;

            // 生成 ns_1-name_1 具名实例的句柄, 可以重用 ns_1-name_1 具名实例
            counter = YR::Instance(Counter::FactoryCreate, "name_1", "ns_1").Invoke(1);
            c = counter.Function(&Counter::Add).Invoke(1);
            std::cout << "counter is " << *YR::Get(c) << std::endl;

            return 0;
         }

    模板参数：
        - **F** - 不能是继承自 `internal::XLangClass` 的类型。

    参数：
        - **constructor** - 一个用于构建类的实例并返回指向所构建对象的指针的函数。
        - **name** - 具名实例的实例名称，可以通过实例名称重用这个实例。
        - **ns** - 具名实例的命名空间。

    返回：
        InstanceCreator<F>：一个创建器对象，可用于创建实例并指定实例资源的额外选项。