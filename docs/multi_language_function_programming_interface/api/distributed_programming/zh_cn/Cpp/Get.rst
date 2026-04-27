Get
======

根据对象的键从后端存储中检索对象的值。接口调用将阻塞，直到检索到对象的值或发生超时。

.. cpp:function:: template<typename T> std::shared_ptr<T>\
                  YR::Get(const ObjectRef<T> &obj, int timeout = DEFAULT_GET_TIMEOUT_SEC)

    获取 ObjectRef 的值。

    根据对象的键从后端存储中获取对象的值。接口调用将阻塞，直到获取到对象的值或发生超时。

    .. code-block:: cpp

       auto objRef = YR::Put(100);
       auto value = *(YR::Get(objRef));
       assert(value == 100);  // 应该为 100

    模板参数：
        - **T** - 对象的类型，不能为 `void`。

    参数：
        - **obj** - 对象的 `ObjectRef`。
        - **timeout** - 超时时间。默认 ``300`` s。
    
    抛出：
        :cpp:class:`Exception` - 如果对象不存在或发生超时。
  
    返回：
        实际值的共享指针。

.. cpp:function:: template<typename T> std::vector<std::shared_ptr<T>>\
                  YR::Get(const std::vector<ObjectRef<T>> &objs, int timeout = DEFAULT_GET_TIMEOUT_SEC, bool allowPartial = false)

    获取一个 `ObjectRef` 的值。

    .. code-block:: cpp

       int originValue = 100;
       auto objRefs = std::vector<YR::ObjectRef<int>>{YR::Put(100), YR::Put(101)};
       auto values = *(YR::Get(objRefs));
       assert(values.size() == 2);  // 应该为 [100, 101]
       assert(*values[0] == 100);
       assert(*values[1] == 101);

    模板参数：
        - **T** - 对象的类型，不能为 `void`。

    参数：
        - **obj** - 对象的 `ObjectRef` 作为一个向量，数量应少于 ``10000``。
        - **timeout** - 超时时间。默认 ``300`` s。
        - **allowPartial** - 如果设置为 ``true``，则部分成功不会抛出错误。
   
    抛出：
        :cpp:class:`Exception` - 如果所有对象都不存在或发生超时，并且不允许部分成功。
 
    返回：
        实际值的共享指针列表。