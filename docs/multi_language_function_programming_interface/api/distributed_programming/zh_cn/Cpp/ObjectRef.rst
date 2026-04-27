ObjectRef
===========

.. cpp:class:: template<typename T> ObjectRef

    对象引用。

    .. note::
        - openYuanrong 鼓励用户使用 `YR::Put <Put.html>`_ 将大对象存储在数据系统中，并获取一个唯一的 `ObjectRef`。在调用用户函数时，使用 `ObjectRef` 而不是对象本身作为函数参数，以减少在 openYuanrong 和用户函数组件之间传输大对象的开销，确保高效的数据流转。

        - 每次用户函数调用的返回值也将以 `ObjectRef` 的形式返回，用户可以将其用作下一次调用的输入参数，或者通过 `YR::Get <Get.html>`_ 获取相应的对象。

        - 目前，用户无法自行构造 `ObjectRef`。

    模板参数：
        - **T** - ObjectRef 的类型。

    **公共函数**
 
    .. cpp:function:: inline std::string ID() const
 
       获得对象 ID。
 
       返回：
           ObjectRef 的 ID。
 
    .. cpp:function:: inline bool IsLocal() const
 
       检查 ObjectRef 是否为本地。
 
       返回：
           如果 ObjectRef 是本地的，则返回 ``true``，否则返回 ``false``。
 
