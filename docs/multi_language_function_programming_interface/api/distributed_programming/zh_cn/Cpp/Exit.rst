Exit
==========

.. cpp:function:: void YR::Exit()

    退出当前函数实例。

    此函数不支持本地调用（无论是在 `LOCAL_MODE` 中，还是在 `CLUSTER_MODE` 中的本地调用）。
    如果在本地调用，将抛出“Not support exit out of cluster”的异常。

    .. code-block:: cpp

       void AddOneAndPut(int x) {
           YR::KV().Set("key", std::to_string(x));
           // 设置键和值后，函数实例将退出。
           YR::Exit();
       }

       YR_INVOKE(AddOneAndPut);

    抛出：
        :cpp:class:`Exception` - 如果在本地环境中调用，它会抛出“Not support exit out of cluster”的异常。