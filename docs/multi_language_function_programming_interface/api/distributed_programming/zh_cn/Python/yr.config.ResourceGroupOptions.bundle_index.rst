.. _bundle_index:

yr.config.ResourceGroupOptions.bundle_index
----------------------------------------------

.. py:attribute:: ResourceGroupOptions.bundle_index
   :type: int
   :value: -1

   资源组名称不为空时，要调度的 bundle 的索引才生效。取值范围是 [-1, 资源组中的 bundle 数量)。
   默认值是 ``-1``，表示不指定具体的 bundle；如果取值范围内的值不是 ``-1``，则表示将 bundle 调度到资源组中的相应索引；如果是其他任何值，将生成错误。