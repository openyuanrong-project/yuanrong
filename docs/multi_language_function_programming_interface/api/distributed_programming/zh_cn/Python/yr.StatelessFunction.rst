yr.StatelessFunction
===================

.. currentmodule:: yr

.. py:class:: StatelessFunction

   无状态函数的代理对象，提供远程执行能力。

   StatelessFunction 是 :py:func:`yr.invoke` 装饰器返回的对象类型，用于表示可以在 Yuanrong 分布式系统上远程执行的无状态函数。

   .. note::
       
      StatelessFunction 是推荐使用的新名称，替代了原来的 FunctionProxy。FunctionProxy 仍然可用以保持向后兼容性。

   **方法**：

   .. list-table::
      :widths: 40 60
      :header-rows: 0

      * - :ref:`invoke <invoke_sf>`
        - 执行被装饰的远程函数调用。
      * - :ref:`options <options_sf>`
        - 动态修改被装饰函数的调用参数。

.. toctree::
   :maxdepth: 1
   :hidden:

   yr.StatelessFunction.invoke
   yr.StatelessFunction.options
