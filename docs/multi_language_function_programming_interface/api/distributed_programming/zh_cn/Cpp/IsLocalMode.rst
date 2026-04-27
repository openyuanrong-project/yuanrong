IsLocalMode
=============

.. cpp:function:: bool YR::IsLocalMode()

    确定 openYuanrong 是否运行在本地模式。


    抛出：
        :cpp:class:`Exception` - 在初始化 openYuanrong 前调用该接口，会抛出“Please init YR first”的异常。

    返回：
        openYuanrong 运行在本地模式返回 ``true``，否则返回 ``false``。
