.. _init_IO:

yr.InvokeOptions.__init__
-----------------------------------

.. py:method:: InvokeOptions.__init__(cpu: int = 500, memory: int = 500, concurrency: int = 1, custom_resources: ~typing.Dict[str, float] = <factory>, custom_extensions: ~typing.Dict[str, str] = <factory>, pod_labels: ~typing.Dict[str, str] = <factory>, labels: ~typing.List[str] = <factory>, max_invoke_latency: int = 5000, min_instances: int = 0, max_instances: int = 0, recover_retry_times: int = 0, need_order: bool = False, name: str = '', namespace: str = '', schedule_affinities: ~typing.List[~yr.affinity.Affinity] = <factory>, resource_group_options: ~yr.config.ResourceGroupOptions = <factory>, function_group_options: ~yr.config.FunctionGroupOptions = <factory>, env_vars: ~typing.Dict[str, str] = <factory>, retry_times: int = 0, trace_id: str = '', alias_params: ~typing.Dict[str, str] = <factory>, runtime_env: ~typing.Dict = <factory>) -> None
