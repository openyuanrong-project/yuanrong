Java
==============================

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1
   
   init
   Finalize
   Config

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   invoke
   function
   FunctionHandler
   function-options
   Instance
   InstanceCreator
   InstanceHandler
   InstanceHandler-function
   InstanceFunctionHandler
   InstanceHandler-exportHandler
   InstanceHandler-importHandler
   InstanceHandler-terminate
   saveState
   loadState
   yrShutdown
   yrRecover
   InvokeOptions

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   put
   get
   wait
   ObjectRef

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   setUrn
   cpp-function
   CppFunctionHandler
   cpp_instance_class
   cpp_instance_method
   CppInstanceCreator
   CppInstanceHandler
   CppInstanceHandler-function
   CppInstanceFunctionHandler
   CppInstanceHandler-exportHandler
   CppInstanceHandler-importHandler
   CppInstanceHandler-terminate
   java-function
   JavaFunctionHandler
   java_instance_class
   java_instance_method
   JavaInstanceCreator
   JavaInstanceHandler
   JavaInstanceHandler-function
   JavaInstanceFunctionHandler
   JavaInstanceHandler-exportHandler
   JavaInstanceHandler-importHandler
   JavaInstanceHandler-terminate
   VoidFunctionHandler
   VoidInstanceFunctionHandler

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   kv.set
   kv.mSetTx
   kv.write
   kv.mWriteTx
   kv.get
   kv.getWithParam
   kv.read
   kv.del

.. toctree::
   :glob:
   :hidden:
   :maxdepth: 1

   Group

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
-----------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`init`
     - The openYuanrong initialization interface is used to configure parameters.
   * - :doc:`Finalize`
     - Terminate the system and release the relevant resources.
   * - :doc:`Config`
     - Configuration parameters required by the init() interface.


Stateful & Stateless Function API
--------------------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`invoke`
     - Create openYuanrong function instance or call method.
   * - :doc:`function`
     - Constructs a FunctionHandler for a given function.
   * - :doc:`FunctionHandler`
     - Create an operation class for creating a stateless Java function instance.
   * - :doc:`function-options`
     - Set the options for the current invocation, including resources, etc.
   * - :doc:`Instance`
     - Create an InstanceCreator for constructing an instance of a class.
   * - :doc:`InstanceCreator`
     - Create an operation class for creating a Java stateful function instance.
   * - :doc:`InstanceHandler`
     - Handle to an instance of a stateful function.
   * - :doc:`InstanceHandler-function`
     - Constructs a InstanceFunctionHandler for a given stateful function instance member method.
   * - :doc:`InstanceFunctionHandler`
     - Handle to a stateful function instance member method.
   * - :doc:`InstanceHandler-exportHandler`
     - Export handler information for stateful function instance.
   * - :doc:`InstanceHandler-importHandler`
     - Import handler information for stateful function instance.
   * - :doc:`InstanceHandler-terminate`
     - Terminate a stateful function instance.
   * - :doc:`saveState`
     - Saves the instance state.
   * - :doc:`loadState`
     - Loads the saved instance state.
   * - :doc:`yrShutdown`
     - Users can use this interface to perform data cleanup or data persistence operations.
   * - :doc:`yrRecover`
     - Users can use this interface to perform data recovery operations.
   * - :doc:`InvokeOptions`
     - Use to set the invoke options.


Data Object API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`put`
     - Save data to the backend.
   * - :doc:`get`
     - Retrieve the value of an object from the backend storage based on its key. 
   * - :doc:`wait`
     - Waits for the values of a set of objects in the data system to be ready based on their keys.
   * - :doc:`ObjectRef`
     - Initialize the ObjectRef.


Function Interoperation API
-----------------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`setUrn`
     - Set the functionUrn for the function.
   * - :doc:`cpp-function`
     - Java function calls C++ function.
   * - :doc:`CppFunctionHandler`
     - Class for creating instances of stateless functions in C++.
   * - :doc:`cpp_instance_class`
     - Helper class to new C++ instance class. 
   * - :doc:`cpp_instance_method`
     - Helper class to invoke member function.
   * - :doc:`CppInstanceCreator`
     - Create an operation class for creating a C++ stateful function instance. 
   * - :doc:`CppInstanceHandler`
     - Handle to an instance of a C++ stateful function.
   * - :doc:`CppInstanceHandler-function`
     - Constructs a InstanceFunctionHandler for a given C++ stateful function instance member method.
   * - :doc:`CppInstanceFunctionHandler`
     - Handle to a C++ stateful function instance member method.
   * - :doc:`CppInstanceHandler-exportHandler`
     - Obtain instance handle information. 
   * - :doc:`CppInstanceHandler-importHandler`
     - Import handler information for C++ stateful function instance.
   * - :doc:`CppInstanceHandler-terminate`
     - Terminate a C++ function instance.
   * - :doc:`java-function`
     - Java function calls Java function.
   * - :doc:`JavaFunctionHandler`
     - Create an operation class for creating a stateless Java function instance.
   * - :doc:`java_instance_class`
     - Helper class to new java instance class.
   * - :doc:`java_instance_method`
     - Helper class to invoke member function.
   * - :doc:`JavaInstanceCreator`
     - Create an operation class for creating a Java stateful function instance.
   * - :doc:`JavaInstanceHandler`
     - Handle to an instance of a Java stateful function. 
   * - :doc:`JavaInstanceHandler-function`
     - Constructs a InstanceFunctionHandler for a given Java stateful function instance member method. 
   * - :doc:`JavaInstanceFunctionHandler`
     - Handle to a Java stateful function instance member method.
   * - :doc:`JavaInstanceHandler-exportHandler`
     - Obtain instance handle information.
   * - :doc:`JavaInstanceHandler-importHandler`
     - Import handler information for Java stateful function instance.
   * - :doc:`JavaInstanceHandler-terminate`
     - Terminate a Java function instance.
   * - :doc:`VoidFunctionHandler`
     - Operation class that calls a void function with no return value.
   * - :doc:`VoidInstanceFunctionHandler`
     - Operation class that calls a Java class instance member function without a return value.


KV Cache API
-----------------

.. list-table::
   :header-rows: 0
   :widths: 30 70

   * - :doc:`kv.set`
     - Sets the value of a key.
   * - :doc:`kv.mSetTx`
     - A transactional interface for setting multiple binary data entries in a batch.
   * - :doc:`kv.write`
     - Writes the value of a key.
   * - :doc:`kv.mWriteTx`
     - Sets multiple key-value pairs.
   * - :doc:`kv.get`
     - Retrieves a value associated with a specified key, similar to Redis’s GET command.
   * - :doc:`kv.getWithParam`
     - Retrieves multiple values associated with specified keys with additional parameters, supporting offset-based reading.
   * - :doc:`kv.read`
     - Retrieve the value of the specified type associated with the specified key.
   * - :doc:`kv.del`
     - Deletes a key and its associated data, similar to Redis’s DEL command.


Function Group API
--------------------

.. list-table::
   :header-rows: 0
   :widths: 30 70
 
   * - :doc:`Group`
     - A class for managing the lifecycle of grouped instances.

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
