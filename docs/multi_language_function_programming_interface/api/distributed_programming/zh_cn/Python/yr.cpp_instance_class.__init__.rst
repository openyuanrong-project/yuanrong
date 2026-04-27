.. _init_cpp:

yr.cpp_instance_class.__init__
---------------------------------

.. py:method:: cpp_instance_class.__init__(class_name: str, factory_name: str, function_urn: str)

    初始化。

    参数：
        - **class_name** (str) – cpp 类名。
        - **factory_name** (str) – cpp 类静态构造函数名。
        - **function_urn** (str, optional) – 函数 urn，默认为 ``sn:cn:yrk:default:function:0-defaultservice-cpp:$latest``。

    样例：

    .. code-block:: cpp

       #include <iostream>
       #include "yr/yr.h"

           class Counter {
               public:

                   int count;
                   Counter() {}
                   Counter(int init) { count = init; }

                   static Counter* FactoryCreate(int init) {
                       std::cout << "Counter FactoryCreate with init=" << init << std::endl;
                       return new Counter(init);
                   }

                   int Add(int x) {
                       count += x;
                       std::cout << "Counter Add with init=" << count << std::endl;
                       return count;
                   }

                   YR_STATE(count);
           };

           YR_INVOKE(Counter::FactoryCreate, &Counter::Add);

    .. code-block:: python

       >>> import yr
       >>> yr.init()
       >>> cpp_function_urn = (
       ...     "sn:cn:yrk:default:"
       ...     "function:0-yr-defaultservice-cpp:$latest"
       ... )
       >>> counter_class = yr.cpp_instance_class("Counter", "Counter::FactoryCreate", cpp_function_urn)
       >>> opt = yr.InvokeOptions(cpu=1000, memory=1024)
       >>> ins = counter_class.options(opt).invoke(11)
       >>> result = ins.Add.invoke(9)
       >>> yr.get(result)
       >>> ins.terminate()
       >>> yr.finalize()
