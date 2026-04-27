PyFunction
===========

.. cpp:class:: template<typename R> PyFunctionHandler : public YR::internal::XlangBaseType

    模板参数：
        - **R** - 返回值的类型。
    
    **公共函数**

    .. cpp:function:: inline R operator()()

.. cpp:function:: template<typename R>\
                  FunctionHandler<PyFunctionHandler<R>> YR::PyFunction(const std::string &moduleName, const std::string &functionName)
    
    用于 C++ 调用 Python 函数，构造对 Python 函数的调用。

    .. code-block:: cpp

       int main(void)
       {
           // def add_one(a):
           //     return a + 1
           YR::Config conf;
           YR::Init(conf);
    
           auto r1 = YR::PyFunction<int>("pycallee", "add_one").Invoke(x);  // moduleName, functionName
           auto res = *YR::Get(r1);
    
           std::cout << "PlusOneWithPyFunc with result=" << res << std::endl;
           return res;
    
           return 0;
       }

    模板参数：
        - **R** - 函数的返回类型。

    参数：
        - **moduleName** - 函数所在的 Python 模块的名称。
        - **functionName** - Python 函数的名称。
  
    返回：
        一个 `FunctionHandler` 对象，提供执行函数的方法。`PyFunctionHandler` 是 `FunctionHandler` 中的一个模板类，可用于获取返回类型。