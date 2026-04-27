PyInstanceClass::FactoryCreate
===============================

.. cpp:function:: static inline PyInstanceClass\
                  YR::PyInstanceClass::FactoryCreate(const std::string &moduleName, const std::string &className)

    创建一个 `PyInstanceClass` 对象，该对象可以作为参数传递给 `YR::Instance <Instance.html>`_，以成功创建一个 Python 函数类实例。

    .. code-block:: cpp

       int main(void)
       {
           // class Instance:
           //     sum = 0
           //
           //     def __init__(self, init):
           //       self.sum = init
           //
           //     def add(self, a):
           //       self.sum += a
           //
           //     def get(self):
           //       return self.sum
           YR::Config conf;
           YR::Init(conf);
    
           auto pyInstance = YR::PyInstanceClass::FactoryCreate("pycallee", "Instance");  // moduleName, className
           auto r1 = YR::Instance(pyInstance).Invoke(x);
           r1.PyFunction<void>("add").Invoke(1);  // returnType, memberFunctionName
    
           auto r2 = r1.PyFunction<int>("get").Invoke();
           auto res = *YR::Get(r2);
    
           std::cout << "PlusOneWithPyClass with result=" << res << std::endl;
           return res;
    
           return 0;
       }

    参数：
        - **moduleName** - Python模块名称。
        - **classname** - Python类名称。
    
    返回：
        一个携带函数信息的 `PyInstanceClass` 对象，该对象可以作为参数传递给 `YR::Instance <Instance.html>`_。