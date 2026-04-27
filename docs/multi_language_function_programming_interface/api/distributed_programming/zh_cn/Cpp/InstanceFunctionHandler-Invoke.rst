InstanceFunctionHandler::Invoke
================================

.. cpp:function:: template<typename ...Args>\
                  inline ObjectRef<ReturnType<F>> YR::InstanceFunctionHandler::Invoke(Args&&... args)

    通过向远程后端发送请求来调用远程函数以执行指定的函数。函数的参数将被序列化并传输到后端，结果将作为 `ObjectRef` 返回。

    该函数支持基于参数类型的依赖关系解析：

    1、如果函数的参数类型为 `ArgType`，而传递的参数类型为 `ObjectRef<ArgType>`，则只有在完成 `obj` 对应的计算后，才会发送请求，并且 `obj` 需来自同一个客户端。
    
    2、如果函数的参数类型为 `vector<ObjectRef<ArgType>> objs`，而传递的参数类型为 `vector<ObjectRef<ArgType>>`，则只有在完成 `objs` 对应的所有计算后，才会发送请求，并且所有 `ObjectRef` 实例都需来自同一个客户端。
    
    3、不支持对其他类型参数的依赖关系解析。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);

           auto ins = YR::Instance(SimpleCaculator::Constructor).Invoke();
           auto r3 = ins.Function(&SimpleCaculator::Plus).Invoke(1, 1);
           int res = *(YR::Get(r3));

           return 0;
       }

    在这个示例中，使用 `Invoke` 方法调用了 `SimpleCaculator` 实例的 `Plus` 函数，参数为 ``1`` 和 ``1``。通过 `YR::Get(r3)` 获取结果。

    模板参数：
        - **Args** - 传递给函数的参数类型。

    参数：
        - **args** - 要传递给远程函数的参数。参数的类型和数量需与函数的定义完全匹配。确保参数类型与预期类型精确匹配，以避免因隐式类型转换导致的问题。
    
    抛出：
        :cpp:class:`Exception` - 如果调用失败，例如由于函数实例异常退出或用户代码执行异常。
  
    返回：
        ObjectRef<ReturnType<F>>：对结果对象的引用，本质上是一个键。要获取实际的值，请使用 `YR::Get <Get.html>`_ 方法。