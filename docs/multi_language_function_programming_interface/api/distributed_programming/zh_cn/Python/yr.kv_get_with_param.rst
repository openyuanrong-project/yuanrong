yr.kv_get_with_param
=====================

.. py:function:: yr.kv_get_with_param(keys: List[str], get_params: GetParams, timeout: int = 300) \
                                      -> List[Union[bytes, memoryview]]

    提供支持偏移读的接口 kv_get_with_param, 支持偏移读取数据。

    参数：
        - **keys** (List[str]) – 查询数据指定的多个键。
        - **get_params** (GetParams_) – 为获取 value 配置一组属性，每组属性包括 `offset（偏移量）`、`size（大小）` 等参数。
        - **timeout** (int) – 单位为秒，默认为 ``300``，为 ``-1`` 时表示永久阻塞等待。参数需大于等于 ``-1``。
    返回：
        返回查到的一个或一组数据。
        数据类型：Union[bytes, List[bytes]]。
    
    异常：
        - **ValueError** - 如果入参 `get_params` 的 `GetParam` 类型列表 `get_params` 列表是空列表。
        - **ValueError** - 如果入参 `get_params` 的 `GetParam` 类型列表 `get_params` 列表个数不等于入参 `keys` 个数。
        - **RuntimeError** – 如果未初始化调用 `kv_get_with_param()`，会抛出异常。
        - **RuntimeError** – 如果从数据系统获取数据失败。
    样例：
        >>> get_param = yr.GetParam()
        >>> get_param.offset = 0
        >>> get_param.size = 0
        >>> params = yr.GetParams()
        >>> params.get_params = [get_param]
        >>> v1 = yr.kv_get_with_param(["kv-key"], params, 10)

.. _GetParams: ../../zh_cn/Python/yr.runtime.GetParams.html#yr.runtime.GetParams