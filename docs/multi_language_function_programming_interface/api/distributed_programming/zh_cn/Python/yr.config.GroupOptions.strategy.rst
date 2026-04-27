.. _strategy:

yr.config.GroupOptions.strategy
------------------------------------

.. py:attribute:: GroupOptions.strategy
   :type: str
   :value: ''

   创建实例组的策略。None：不使用策略。SPREAD：尽可能将多个实例分散到不同节点上。STRICT_PACK：所有实例必须放置在同一节点，否则创建失败。PACK：尽可能将多个实例打包到同一节点。STRICT_SPREAD：所有实例必须放置在不同节点，否则创建失败。默认值：``None``。