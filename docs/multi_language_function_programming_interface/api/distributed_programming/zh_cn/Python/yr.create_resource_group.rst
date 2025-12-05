yr.create_resource_group
==========================

.. py:function:: yr.create_resource_group(bundles: List[Dict[str, float]], name: Optional[str] = None, strategy: Optional[str] = "PACK") -> ResourceGroup

    异步创建一个 ResourceGroup。 

    .. note::
        - 其中 `bundle_index` 为各个 `bundle` 在列表 `bundles` 中的下标。
        - 合法的资源名包括："CPU" 、 "Memory" 和自定义资源。其中自定义资源当前支持 "GPU/XX/YY" 和 "NPU/XX/YY" ，"XX" 为
          卡的型号如 "Ascend910B4" ，"YY" 可以是 "count" , "latency" , "stream" 。
        - 各个 `bundle` 中的 "CPU" 、 "Memory" 若未设置，或者置为 ``0`` ，则会分别被赋值 ``300`` 和 ``128`` 。
        - 各个 `bundle` 中的资源量需要大于 ``0`` ，若小于 ``0`` 则抛出异常；资源名非空，若为空则抛出异常。
        - 除 "CPU" 、 "Memory" 外的其他资源若被置为 ``0`` ，则会被过滤。

    参数：
        - **bundles** (List[Dict[str, float]]) - 一组代表资源请求的 `bundles`，不能为空。
        - **name** (Optional[str]，可选) - 待创建 ResourceGroup 的名称，具有唯一性，且不能是 'primary' 或者空字符串。
          该参数可选，默认值为 ``None`` ，即随机生成 rgroup-{uuid} 类型的字符串作为 `resource group name` 。
        - **strategy** (Optional[str]，可选) - 创建 ResourceGroup 的策略，默认策略为 ``'PACK'`` 。
          
          - ``'NONE'``：无策略。

          - ``'PACK'``：尽可能将多个 `bundle`` 集中在同一个节点。
          
          - ``'SPREAD'``：尽可能将多个 `bundle` 分散到不同节点。
          
          - ``'STRICT_PACK'``：所有 `bundle` 需放在同一节点，否则创建失败。
          
          - ``'STRICT_SPREAD'``：所有 `bundle` 需放在不同节点，否则创建失败。
    返回：
        一个 ResourceGroup_ 句柄。

    异常：
        - **TypeError** – 参数类型有误。
        - **ValueError** – 参数不能为空。
        - **RuntimeError** – 创建 resource group 资源不足。
        - **RuntimeError** – resource group 重复创建。
        - **RuntimeError** – resource group 名称不合法。
	
    样例：
        >>> rg1 = yr.create_resource_group([{"NPU/Ascend910B4/count":1},{"CPU":2000,"Memory":2000}])
        >>>
        >>> rg2 = yr.create_resource_group([{"NPU/Ascend910B4/count":1},{"CPU":2000,"Memory":2000}], "rgname")

.. _ResourceGroup: ../../zh_cn/Python/yr.ResourceGroup.html#yr.ResourceGroup