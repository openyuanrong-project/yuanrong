.. _resource_group_options:

yr.InvokeOptions.resource_group_options
-------------------------------------------

.. py:attribute:: InvokeOptions.resource_group_options
   :type: ResourceGroupOptions

   指定 ResourceGroup 选项，包括 resource_group_name 和 bundle_index。

   当创建函数实例时：

     - 如果设置了 resource_group_name，则会将其传递给内核，以将实例调度到指定的 ResourceGroup。
     - 如果同时设置了 resource_group_name 和 bundle_index，则会将它们传递给内核，以将实例调度到指定的 ResourceGroup 和索引。resource_group_name 的默认值为空，bundle_index 的默认值为 -1。
   
   约束：

     - 当 resource_group_name 为空时，实例不会被调度到指定的 ResourceGroup，bundle_index 字段无效。
     - 当 resource_group_name 不为空时：
     - 当 bundle_index 为 -1 时，实例被调度到指定的 ResourceGroup。
     - 当 0 <= bundle_index < ResourceGroup 中的 bundle 数量时，实例被调度到指定 ResourceGroup 中的指定 bundle。
     - 当 bundle_index < -1 或 bundle_index >= ResourceGroup 中的 bundle 数量时，将引发错误。
