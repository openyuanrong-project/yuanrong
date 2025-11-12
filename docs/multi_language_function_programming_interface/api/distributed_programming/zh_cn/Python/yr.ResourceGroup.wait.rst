.. _wait_ResourceGroup:

yr.ResourceGroup.wait
---------------------------------------------------

.. py:method:: ResourceGroup.wait(timeout_seconds: int | None = None)-> None

    阻塞并等待创建 ResourceGroup 的结果。

    参数：
        - **timeout_seconds** (int, 可选) - 在等待创建 ResourceGroup 结果时阻塞的超时时间。
          默认值为 ``None``，即 ``-1``（从不超时）。取值范围为整数大于等于 ``-1``。如果它等于 ``-1``，它将无限期地等待。

    异常：
        - **ValueError** - 超时时间小于 ``-1``。

    样例：
        >>> rg = yr.create_resource_group([{"NPU":1},{"CPU":2000,"Memory":2000}], "rgname")
        >>> rg.wait(60)
