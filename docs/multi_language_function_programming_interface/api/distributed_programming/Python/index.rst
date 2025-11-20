Python
==============================

.. TODO yr_shutdown  generator 

Basic API
---------

.. autosummary::
   :nosignatures:
   :recursive:
   :toctree: generated/

   yr.init
   yr.is_initialized
   yr.finalize
   yr.Config
   yr.config.ClientInfo


Stateful & Stateless Function API
-----------------------------------

.. autosummary::
   :nosignatures:
   :recursive:
   :toctree: generated/

   yr.invoke
   yr.FunctionProxy
   yr.instance
   yr.InstanceProxy
   yr.method
   yr.MethodProxy
   yr.get_instance
   yr.InstanceCreator
   yr.cancel
   yr.exit
   yr.save_state
   yr.load_state
   yr.InvokeOptions
   yr.list_named_instances


Data Object API
-------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.put
   yr.get
   yr.wait
   yr.object_ref.ObjectRef


Stream API
-----------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.ProducerConfig
   yr.SubscriptionConfig
   yr.Element
   yr.create_stream_producer
   yr.create_stream_consumer
   yr.delete_stream
   yr.fnruntime.Producer
   yr.fnruntime.Consumer

   
Function Interoperation API
-----------------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.cpp_function
   yr.cpp_instance_class
   yr.java_function
   yr.java_instance_class


Function Group API
-------------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.create_function_group
   yr.get_function_group_context
   yr.FunctionGroupOptions
   yr.FunctionGroupContext
   yr.FunctionGroupHandler
   yr.FunctionGroupMethodProxy
   yr.device.DataInfo


Resource Group API
----------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.create_resource_group
   yr.remove_resource_group
   yr.ResourceGroup
   yr.config.ResourceGroupOptions
   yr.config.SchedulingAffinityType
   yr.config.UserTLSConfig
   yr.config.DeploymentConfig



KV Cache API
-------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.kv_write
   yr.kv_write_with_param
   yr.kv_m_write_tx
   yr.kv_read
   yr.kv_del
   yr.kv_set
   yr.kv_get
   yr.kv_get_with_param
   yr.runtime.ExistenceOpt
   yr.runtime.WriteMode
   yr.runtime.CacheType
   yr.runtime.ConsistencyType
   yr.runtime.GetParam
   yr.runtime.GetParams
   yr.runtime.SetParam
   yr.runtime.MSetParam
   yr.runtime.CreateParam


Observability API
--------------------

.. autosummary::
   :nosignatures:
   :toctree: generated/

   yr.Gauge
   yr.Alarm
   yr.AlarmInfo
   yr.AlarmSeverity
   yr.DoubleCounter
   yr.UInt64Counter
   yr.resources


Affinity Scheduling
----------------------

.. autosummary::
   :nosignatures:
   :recursive:
   :toctree: generated/
   
   yr.affinity.AffinityType
   yr.affinity.AffinityKind
   yr.affinity.OperatorType
   yr.affinity.LabelOperator
   yr.affinity.Affinity


Exceptions
---------------------

.. autosummary::
   :nosignatures:
   :recursive:
   :toctree: generated/

   yr.exception.YRError
   yr.exception.CancelError
   yr.exception.YRInvokeError
   yr.exception.YRequestError
