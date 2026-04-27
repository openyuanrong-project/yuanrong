.. _invoke_func:

yr.Function.invoke
-----------------------

.. py:method:: Function.invoke(payload: Union[str, dict] = None) -> ObjectRef

    调用函数。

    参数：
        - **payload** (Union[str, dict]) - 被调用函数的参数。

    返回：
        ObjectRef：对象引用。

    样例：
        >>> from functionsdk import Function, InvokeOptions
        >>> def my_handler(event, context)
        >>>     f = Function("hello", context)
        >>>     objRef = f.invoke(event)
        >>>     res = objRef.get()
        >>>     return {
        >>>        "statusCode": 200,
        >>>        "isBase64Encoded": False,
        >>>        "body": res,
        >>>        "headers": {
        >>>             "Content-Type": "application/json"
        >>>         }
