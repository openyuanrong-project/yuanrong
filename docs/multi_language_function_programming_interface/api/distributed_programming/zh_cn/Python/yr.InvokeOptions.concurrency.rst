.. _concurrency_IO:

yr.InvokeOptions.concurrency
--------------------------------

.. py:attribute:: InvokeOptions.concurrency
   :type: int
   :value: 1

   实例并发度。取值范围为 [1, 1000]。优先级高于在 `custom_extensions` 中配置的“Concurrency”。建议使用此参数进行配置。