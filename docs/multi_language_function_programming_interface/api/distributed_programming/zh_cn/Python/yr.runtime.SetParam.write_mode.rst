.. _write_mode_sp:

yr.runtime.SetParam.write_mode
------------------------------------

.. py:attribute:: SetParam.write_mode
   :type: WriteMode
   :value: 0

   设置数据的可靠性。当服务端配置支持二级缓存时用于保证可靠性时，比如 redis 服务，那么通过这个配置可以保证数据可靠性。
   默认值为 ``WriteMode.NONE_L2_CACHE``，即 ``0``。