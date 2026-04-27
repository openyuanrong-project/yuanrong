Cancel
=============

.. cpp:function:: template<typename T>\
                  void YR::Cancel(const ObjectRef<T> &obj, bool isForce = true, bool isRecursive = false)

    通过指定的 ObjectRef 取消对应的函数调用。

    .. code-block:: cpp

       auto obj = YR::Function(Handler).Invoke(1);
       YR::Cancel(obj);

    模板参数：
        - **T** - 取消的对象类型。

    参数：
        - **obj** - 数据系统中对象引用的集合。
        - **isForce** - 设置为 ``true`` 时，如果当前正在执行函数调用，则会强制终止与该函数调用对应的函数实例进程。注意：目前不支持强制终止并发配置不为 1 的函数实例。默认值为 ``true``。
        - **isRecursive** - 设置为 ``true`` 将取消嵌套函数调用。默认值为 ``false``。

.. cpp:function:: template<typename T>\
                  void YR::Cancel(const std::vector<ObjectRef<T>> &objs, bool isForce = true, bool isRecursive = false)

    通过指定一组 ObjectRef 来取消对应的函数调用。

    .. code-block:: cpp

       int num = 5;
       std::vector<YR::ObjectRef<int>> vec;
       for (int i = 0; i < num; ++i) {
           auto obj = YR::Function(Handler).Invoke(i);
           vec.emplace_back(std::move(obj));
       }
       YR::Cancel(vec);

    模板参数：
        - **T** - 取消的对象类型。

    参数：
        - **objs** - 数据系统中对象引用的集合。
        - **isForce** - 设置为 ``true`` 时，如果当前正在执行函数调用，则会强制终止与该函数调用对应的函数实例进程。注意：目前不支持强制终止并发配置不为 1 的函数实例。默认值为 ``true``。
        - **isRecursive** - 设置为 ``true`` 将取消嵌套函数调用。默认值为 ``false``。

    抛出：
        :cpp:class:`Exception` - 
        
        - 本地模式不支持取消操作，将会抛出“local mode does not support cancel”的异常。
        - objs 不能为空，否则会抛出异常“Cancel does not accept empty object list”。

    