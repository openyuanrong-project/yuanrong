.. _getUserData:

yr.Context.getUserData
-----------------------

.. py:method:: Context.getUserData(key, default=None)

    获取用户通过环境变量传入的值。

    参数：
        - **key** (string) - 用户配置的环境变量的键。
        - **default** (string) - 用户获取到空环境变量时的默认值。

    返回：
        用户配置的环境变量中对应键的值。
