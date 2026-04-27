.. _custom_resources:

yr.InvokeOptions.custom_resources
----------------------------------------------

.. py:attribute:: FunctionGroupOptions.custom_resources
   :type: Dict[str, float]

   自定义资源目前支持“GPU/XX/YY”和“NPU/XX/YY”，其中 XX 是卡型号，如 Ascend910B4，YY 可以是 count、latency 或 stream。