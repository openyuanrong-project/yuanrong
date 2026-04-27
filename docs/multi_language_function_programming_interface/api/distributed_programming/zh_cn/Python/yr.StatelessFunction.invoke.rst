yr.StatelessFunction.invoke
==========================

.. currentmodule:: yr

.. py:method:: yr.StatelessFunction.invoke(*args, **kwargs)

   执行被装饰的远程函数调用。

   此方法用于远程执行通过 :py:func:`yr.invoke` 装饰器装饰的函数。调用会立即返回一个 :py:class:`ObjectRef` 对象，该对象表示远程函数执行的结果引用。

   :参数:
       - **\*args** -- 传递给被装饰函数的位置参数
       - **\*\*kwargs** -- 传递给被装饰函数的关键字参数

   :返回:
       远程函数执行结果的对象引用。数据类型为 :py:class:`ObjectRef`。

   :异常:
       - **RuntimeError** -- 如果远程函数执行失败或系统未初始化
       - **TypeError** -- 如果参数类型不正确

   .. rubric:: 示例

   基本用法::

       import yr
       yr.init()

       @yr.invoke
       def add(a, b):
           return a + b

       # 远程调用函数
       result_ref = add.invoke(1, 2)
       result = yr.get(result_ref)
       print(result)  # 输出: 3

       yr.finalize()

   并行执行多个任务::

       import yr
       yr.init()

       @yr.invoke
       def square(x):
           return x * x

       # 并行执行多个任务
       refs = [square.invoke(i) for i in range(5)]
       results = yr.get(refs)
       print(results)  # 输出: [0, 1, 4, 9, 16]

       yr.finalize()

   .. seealso::

      - :py:func:`yr.invoke` - 创建无状态函数装饰器
      - :py:func:`yr.get` - 获取远程执行结果
      - :py:class:`ObjectRef` - 远程对象引用
