Exception
===========

.. cpp:class:: Exception

    openYuanrong 异常类。

    **公共函数**
 
    .. cpp:function:: inline const char *what() const noexcept override
 
       获取异常详细信息，包括错误码，模块码和错误信息。
 
       返回：
           异常详细信息。
 
    .. cpp:function:: inline int Code() const
 
       获取错误码。
 
       返回：
           错误码。

    .. cpp:function:: inline int MCode() const

       获取模块码。

       返回：
           模块码。

    .. cpp:function:: inline std::string Msg() const

       获取错误信息。

       返回：
           错误信息。