IsInitialized
================

.. cpp:function:: bool YR::IsInitialized()

    确定 openYuanrong 是否已初始化。

    .. code-block:: cpp

       int main(void) {
           YR::Config conf;
           YR::Init(conf);
           bool IsInitialized = YR::IsInitialized(); // true

           return 0;
       }

    返回：
        已初始化 openYuanrong 返回 ``true``，否则返回 ``false``。
