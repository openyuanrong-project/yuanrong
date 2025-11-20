C++
==============================

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Init
   IsInitialized
   IsLocalMode
   Finalize
   struct-Config
   struct-InstanceRange
   Exception

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   YR_INVOKE
   YR_STATE
   YR_RECOVER
   YR_SHUTDOWN
   Function
   FunctionHandler-Invoke
   FunctionHandler-Options
   FunctionHandler-SetUrn
   Instance
   InstanceCreator-Invoke
   InstanceCreator-Options
   InstanceCreator-SetUrn
   InstanceFunctionHandler-Invoke
   InstanceFunctionHandler-Options
   NamedInstance
   GetInstance
   Cancel
   Exit
   SaveState
   LoadState
   struct-InvokeOptions

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Put
   Get
   Wait
   ObjectRef

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   PyFunction
   PyInstanceClass-FactoryCreate
   JavaFunction
   JavaInstanceClass-FactoryCreate
   CppFunction
   CppInstanceClass-FactoryCreate

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   struct-Group

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   KV-Set
   KV-MSetTx
   KV-Write
   KV-MWriteTx
   KV-WriteRaw
   KV-Get
   KV-GetWithParam
   KV-Read
   KV-ReadRaw
   KV-Del

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   ParallelFor

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Affinity

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Stream

Basic API
----------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Init`
     - Configures runtime modes and system parameters.
   * - :doc:`IsInitialized`
     - Determine if openYuanrong has been initialized.
   * - :doc:`IsLocalMode`
     - Determine if openYuanrong is in local mode.
   * - :doc:`Finalize`
     - Finalizes the openYuanrong system.
   * - :doc:`struct-Config`
     - Configuration parameters required by the init() interface.
   * - :doc:`struct-InstanceRange`
     - Number of function instances Range configuration parameters.
   * - :doc:`Exception`
     - The exception thrown by openYuanrong.


Stateful & Stateless Function API
------------------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`YR_INVOKE`
     - Register functions for openYuanrong distributed invocation.
   * - :doc:`YR_STATE`
     - Marks a class parameter as state, and openYuanrong will automatically save and recover the state member variables of the class.
   * - :doc:`YR_RECOVER`
     - Users can use this interface to perform data recovery operations.
   * - :doc:`YR_SHUTDOWN`
     - Users can use this interface to perform data cleanup or data persistence operations.
   * - :doc:`Function`
     - Constructs a FunctionHandler for a given function.
   * - :doc:`FunctionHandler-Invoke`
     - Invokes the registered function with the provided arguments.
   * - :doc:`FunctionHandler-Options`
     - Set the options for the current invocation, including resources, etc.
   * - :doc:`FunctionHandler-SetUrn`
     - Set the URN for the current function invocation, which can be used in conjunction with CppFunction or JavaFunction.
   * - :doc:`Instance`
     - Create an InstanceCreator for constructing an instance of a class.
   * - :doc:`InstanceCreator-Invoke`
     - Execute the creation of an instance and construct an object of the class.
   * - :doc:`InstanceCreator-Options`
     - Set options for the instance creation, including resource usage settings.
   * - :doc:`InstanceCreator-SetUrn`
     - Set the function URN for the instance creation.
   * - :doc:`InstanceFunctionHandler-Invoke`
     - Invokes a remote function by sending the request to a remote backend for execution.
   * - :doc:`InstanceFunctionHandler-Options`
     - Sets options for the function invocation, such as timeout and retry count.
   * - :doc:`NamedInstance`
     - Named instance that can invoke the Function method of this class to construct member functions of the instance’s class.
   * - :doc:`GetInstance`
     - Get instance associated with the specified name and nameSpace within a timeout.
   * - :doc:`Cancel`
     - Cancel the corresponding function call by specifying an ObjectRef.
   * - :doc:`Exit`
     - Exit the current function instance.
   * - :doc:`SaveState`
     - Saves the instance state.
   * - :doc:`LoadState`
     - Loads the saved instance state.
   * - :doc:`struct-InvokeOptions`
     - Use to set the invoke options.


Data Object API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Put`
     - Save data to the backend.
   * - :doc:`Get`
     - Retrieve the value of an object from the backend storage based on its key. 
   * - :doc:`Wait`
     - Waits for the values of a set of objects in the data system to be ready based on their keys.
   * - :doc:`ObjectRef`
     - Object reference, which is the key of a data object.


Function Interoperation API
------------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`PyFunction`
     - Used for C++ to call Python functions, constructs a call to a Python function.
   * - :doc:`PyInstanceClass-FactoryCreate`
     - Creates a ``PyInstanceClass`` object. 
   * - :doc:`JavaFunction`
     - Create a ``FunctionHandler`` for invoking a Java function from C++.
   * - :doc:`JavaInstanceClass-FactoryCreate`
     - Create a ``JavaInstanceClass`` object for invoking Java functions.
   * - :doc:`CppFunction`
     - Create a ``FunctionHandler`` for invoking a C++ function by name.
   * - :doc:`CppInstanceClass-FactoryCreate`
     - Create a ``CppInstanceClass`` object for invoking C++ functions. 


Function Group API
---------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`struct-Group`
     - A class for managing the lifecycle of grouped instances.


KV Cache API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`KV-Set`
     - Sets the value of a key.
   * - :doc:`KV-MSetTx`
     - A transactional interface for setting multiple binary data entries in a batch.
   * - :doc:`KV-Write`
     - Writes the value of a key.
   * - :doc:`KV-MWriteTx`
     - Sets multiple key-value pairs.
   * - :doc:`KV-WriteRaw`
     - Writes the value (raw bytes) of a key.
   * - :doc:`KV-Get`
     - Retrieves a value associated with a specified key, similar to Redis’s GET command.
   * - :doc:`KV-GetWithParam`
     - Retrieves multiple values associated with specified keys with additional parameters, supporting offset-based reading.
   * - :doc:`KV-Read`
     - Retrieves the value of a key.
   * - :doc:`KV-ReadRaw`
     - Retrieves the value of a key written by WriteRaw.
   * - :doc:`KV-Del`
     - Deletes a key and its associated data, similar to Redis’s DEL command.


Multithreaded Parallel API
-----------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`ParallelFor`
     - ParallelFor is a function framework for parallel computing, enabling tasks to be executed in parallel across multiple threads to improve computational efficiency.


Affinity Scheduling
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Affinity`
     - Affinity Scheduling Configuration Parameters.

Stream API
-----------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`Stream`
     - Passing ordered unbounded data between openYuanrong functions.  