yr.remove_resource_group
==========================

.. py:function:: yr.remove_resource_group(resource_group: Union[str, ResourceGroup])

    异步删除一个 ResourceGroup。

    参数:
        - **resource_group** (Union[str, ResourceGroup_]) - 待删除 ResourceGroup 的名称或者创建 ResourceGroup 返回的句柄。

    异常:
        - **TypeError** - 入参非 str 或者 ResourceGroup 类型。
        - **RuntimeError** - ResourceGroup 名称不合法。

    样例：
        >>> yr.remove_resource_group("rgname")
        >>>
        >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
        >>> yr.remove_resource_group(rg)

.. _ResourceGroup: ../../zh_cn/Python/yr.ResourceGroup.html#yr.ResourceGroup