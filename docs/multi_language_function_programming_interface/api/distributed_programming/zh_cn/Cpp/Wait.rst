Wait
======

.. cpp:type:: template<typename T>\
              YR::WaitResult = std::pair<std::vector<ObjectRef<T>>, std::vector<ObjectRef<T>>>
    
    表示一个 `vector<ObjectRef>` 对，前者表示已完成的 ObjectRef，后者表示尚未完成的 ObjectRef。

    模板参数：
        - **T** - 正在等待的对象的类型。

.. cpp:function:: template<typename T>\
                  void YR::Wait(const ObjectRef<T> &obj, int timeoutSec = -1)
    
    根据对象的键，在数据系统中等待对象的值准备就绪。

    .. code-block:: cpp

       int timeout = 30;
       auto obj = YR::Function(Handler).Invoke(1);
       YR::Wait(obj, timeout);

    模板参数：
        - **T** - 正在等待的对象的类型。

    参数：
        - **obj** - 数据系统中对象的引用。
        - **timeoutSec** - 超时限制，以毫秒为单位。``-1`` 表示无时间限制。

.. cpp:function:: template<typename T>\
                  WaitResult<T> YR::Wait(const std::vector<ObjectRef<T>> &objs,\
                  std::size_t waitNum, int timeoutSec = -1)
    
    根据对象的键，在数据系统中等待对象的值准备就绪。

    .. code-block:: cpp

       int num = 5;
       std::size_t waitNum = 1;
       std::vector<YR::ObjectRef<int>> vec;
       for (int i = 0; i < num; ++i) {
           auto obj = YR::Function(Handler).Invoke(i);
           vec.emplace_back(std::move(obj));
       }
       int timeout = 30;
       auto waitResult = YR::Wait(vec, waitNum, timeout);
       std::cout << waitResult.first.size() << std::endl;
       std::cout << waitResult.second.size() << std::endl;
 
    模板参数：
        - **T** - 正在等待的对象的类型。
 
    参数：
        - **obj** - 数据系统中对象的引用。
        - **waitnum** - 要等待的最小 `ObjectRef` 数量。
        - **timeoutSec** - 超时限制，以毫秒为单位。``-1`` 表示无时间限制。
  
    返回：
        一个 `vector<ObjectRef>` 对，第一个向量包含已完成的 `ObjectRef`，第二个向量包含尚未完成的 `ObjectRef`。以下不变式始终成立：

        1. waitNum <= waitResult.first.size() <= objs.size()
    
        2. waitResult.first.size() + waitResult.second.size() == objs.size()
