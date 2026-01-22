C++ SDK
===========

.. cpp:class:: Runtime

    openYuanrong 运行时类。

    **公共函数**
 
    .. cpp:function:: Runtime()
 
       默认构造函数。
 
    .. cpp:function:: virtual ~Runtime()
 
       默认析构函数。

    .. cpp:function:: void RegisterHandler(std::function<std::string(const std::string &request, Function::Context &context)> handleRequestFunc）

       向 openYuanrong C++ 运行时注册初始处理函数。

       .. code-block:: cpp

          #include <cstdlib>
          #include <string>
          #include "Runtime.h"
          #include "Function.h"

          bool flags = false;

          std::string HandleRequest(const std::string &request, Function::Context &context)
          {
              Function::FunctionLogger logger = context.GetLogger();
              logger.setLevel("INFO");
              logger.Info("hello cpp %s ", "user info log");
              logger.Error("hello cpp %s ", "user error log");
              logger.Warn("hello cpp %s ", "user warn log");
              logger.Debug("hello cpp %s ", "user debug log");
              logger.Error("hello cpp %s ", context.GetFunctionName().c_str());
              logger.Error("hello cpp %s ", context.GetUserData("b").c_str());
              logger.Error("hello cpp %s ", context.GetUserData("key1").c_str());
              return request;
          }

          void InitState(const std::string &request, Function::Context &context)
          {
              context.SetState(request);
          }

          void Initializer(Function::Context &context)
          {
              flags = true;
          }

          const std::string DEFAULT_PORT = "31552";
          int main(int argc, char *argv[])
          {
              Function::Runtime rt;
              rt.RegisterHandler(HandleRequest); 
              // 初始化函数入口
              rt.RegisterInitializerFunction(Initializer);
              rt.Start(argc, argv);
              return 0;
          }


       参数：
           - **handleRequestFunc** - 待注册的用户自定义处理函数。

    .. cpp:function:: void RegisterInitializerFunction(std::function<void(Function::Context &context)> initializerFunc)

       向 openYuanrong C++ 运行时注册一个初始化处理函数。

       参数：
           - **initializerFunc** - 待注册的用户自定义初始化处理函数。
    
    .. cpp:function:: void RegisterPreStopFunction(std::function<void(Function::Context &context)> preStopFunc)

       向 openYuanrong C++ 运行时注册一个 preStop 处理函数。

       参数：
           - **preStopFunc** - 要注册的用户自定义 preStop 处理程序。

    .. cpp:function:: void Start(int argc, char *argv[])

        启动 openYuanrong C++ 运行时.

        参数：
            - **argc** - 命令行参数的个数。
            - **argv** - 命令行参数字符串数组（其中 argv[0] 为可执行文件名）。

.. cpp:class:: RuntimeHandler

    用户通过实现该接口来处理运行时消息。

    **公共函数**
 
    .. cpp:function:: virtual std::string HandleRequest(const std::string &request, Context &context) = 0
 
       处理传入的运行时请求并返回响应。

       参数：
           - **request** - 传入的请求字符串（通常为 JSON/字符串格式）。
           - **context** - 运行时上下文对象（包含追踪 ID、调用 ID、日志工具）。

       返回：
           std::string，响应请求的字符串。

    .. cpp:function:: virtual void PreStop(Context &context) = 0

       在运行时关闭前执行预停止清理操作。
       当运行时即将关闭时（例如，接收到 SIGINT/SIGTERM 信号）触发此方法。
       该方法应释放资源（例如，关闭文件句柄、数据库连接）并记录清理状态以便进行故障排查。

       参数：
           - **context** - 运行时上下文对象（用于记录关闭状态）。
    
    .. cpp:function:: virtual void Initializer(Context &context) = 0

       执行处理器的全局初始化。
       在运行时启动期间（在 Start() 完成之前）调用此方法一次，用于初始化全局资源（例如，加载配置文件、初始化日志记录器）。
       它在处理任何请求之前执行。

       参数：
           - **context** - 运行时上下文对象（包含运行时配置）。

.. cpp:class:: Context

    这是执行函数时传入的上下文参数类，该对象包含了调用、函数及环境等相关信息。

    **公共函数**
 
    .. cpp:function:: Context() = default
 
       默认构造函数。

       .. code-block:: cpp

          Function::FunctionLogger logger = context.GetLogger();
          logger.setLevel("INFO");
          logger.Info("hello cpp %s ", "user info log");
          std::string envValue = context.GetUserData("a");
          std::string funcName = context.GetFunctionName();
          logger.Info("env: a's value is %s, function name is %s", envValue, funcName)

    .. cpp:function:: virtual ~Context() = default
 
       默认析构函数。
    
    .. cpp:function:: virtual const std::string GetTraceId() const = 0
 
       获取当前处理请求的 traceId。

       返回：
           traceId。

    .. cpp:function:: virtual const FunctionLogger &GetLogger() = 0
 
       获取上下文提供的 FunctionLogger 类实例的引用，用于打印日志。
       具体用法请参考 FunctionLogger 的使用说明。

       返回：
           FunctionLogger 类的实例。
    
    .. cpp:function:: virtual const std::string GetInstanceLabel() const = 0
 
       获取实例标签。

       返回：
           实例标签。

    .. cpp:function:: virtual const std::string GetRequestID() const = 0
 
       获取当前处理请求的 requestId。

       返回：
           requestId。
    
    .. cpp:function:: virtual const std::string GetUserData(std::string key) const = 0
 
       获取用户设置的环境变量。

       参数：
           - **key** - 环境变量的键。

       返回：
           环境变量的值。
    
    .. cpp:function:: virtual const std::string GetFunctionName() const = 0

       获取函数的名称。

       返回：
           函数名称。
    
    .. cpp:function:: virtual const std::string GetVersion() const = 0

       获取函数的版本号。

       返回：
           函数版本。

    .. cpp:function:: virtual int GetMemorySize() const = 0

       获取内存大小（单位：MB）。

       返回：
           内存大小。
    
    .. cpp:function:: virtual int GetCPUNumber() const = 0

       获取 CPU 大小（单位：mi/毫核）。

       返回：
           CPU 大小。
    
    .. cpp:function:: virtual const std::string GetProjectID() const = 0

       获取函数的 tenantId。

       返回：
           tenantId。
    
    .. cpp:function:: virtual const std::string GetPackage() const = 0

       获取函数的服务名称。

       返回：
           服务名称。

.. cpp:class:: FunctionLogger

    日志类。

    **公共函数**
 
    .. cpp:function:: void setLevel(const std::string &level)
 
       设置日志持久化级别。可选级别包括 `DEBUG`、 `INFO`、 `WARN` 和 `ERROR`，默认级别为 `INFO`。

       参数：
           - **level** - 日志级别参数。

       .. code-block:: cpp

          Function::FunctionLogger logger = context.GetLogger();
          logger.setLevel("INFO");
          logger.Info("hello cpp %s ", "user info log");
          logger.Error("hello cpp %s ", "user error log");
          logger.Warn("hello cpp %s ", "user warn log");
          logger.Debug("hello cpp %s ", "user debug log");
          logger.Error("hello cpp %s ", context.GetFunctionName().c_str());
          logger.Error("hello cpp %s ", context.GetUserData("b").c_str());
          logger.Error("hello cpp %s ", context.GetUserData("key1").c_str());
    
    .. cpp:function:: void Info(std::string message, ...)
 
       以 `INFO` 级别打印日志。

       参数：
           - **message** - 日志的格式化字符串。
           - **...** - 变长参数列表，依次对应格式化字符串中的占位符。

    .. cpp:function:: void Warn(std::string message, ...)
 
       以 `WARN` 级别打印日志。

       参数：
           - **message** - 日志的格式化字符串。
           - **...** - 变长参数列表，依次对应格式化字符串中的占位符。

    .. cpp:function:: void Debug(std::string message, ...)
 
       以 `DEBUG` 级别打印日志。

       参数：
           - **message** - 日志的格式化字符串。
           - **...** - 变长参数列表，依次对应格式化字符串中的占位符。

    .. cpp:function:: void Error(std::string message, ...)
 
       以 `ERROR` 级别打印日志。

       参数：
           - **message** - 日志的格式化字符串。
           - **...** - 变长参数列表，依次对应格式化字符串中的占位符。

.. cpp:class:: Function

    此类提供了函数互调功能。

    **公共函数**
 
    .. cpp:function:: explicit Function(Context &context)
 
       Function 类的构造函数，其中被调用方函数为函数本身。

       参数：
           - **context** - Context 对象类型的参数。

    .. cpp:function:: explicit Function(Context &context, const std::string &funcName)
 
       Function 类的构造函数。

       参数：
           - **context** - Context 对象类型的参数。
           - **funcName** - 被调用方函数的名称。

       .. code-block:: cpp

          #include <cstdlib>
          #include <string>
          #include "Runtime.h"
          #include "FunctionError.h"
          #include "Function.h"
          bool flags = false;
          std::string HandleRequest(const std::string &request, Function::Context &context)
          {
              Function::FunctionLogger logger = context.GetLogger();
     
              logger.Info("invoke test function begin");
              std::string result = "";
              try {
                  auto func = Function::Function(context, "test:latest");
                  Function::InvokeOptions invokeOptions; // use default option
                  func.Options(invokeOptions);
                  auto ref = func.Invoke("{\"hello\":\"world\"}");
                  result = ref.Get();
              } catch (Function::FunctionError e) {
                  logger.Error("invoke test function failed, err: %s", e.GetJsonString());
                  return e.GetJsonString();
              }
              logger.Info("invoke test function end, result: %s", result);
     
              return result;
          }
     
          void Initializer(Function::Context &context)
          {
             flags = true;
          }
     
          const std::string DEFAULT_PORT = "31552";
          int main(int argc, char *argv[])
          {
              Function::Runtime rt;
              rt.RegisterHandler(HandleRequest);
             // 初始化函数入口
             rt.RegisterInitializerFunction(Initializer);
             rt.Start(argc, argv);
             return 0;
          }

    .. cpp:function:: virtual ~Function() = default
 
       默认析构函数。

    .. cpp:function:: Function(const Function &) = delete
 
       禁用拷贝构造函数。
    
    .. cpp:function:: Function &operator=(const Function &) = delete
 
       禁用拷贝赋值运算符。

    .. cpp:function:: ObjectRef Invoke(const std::string &payload)
 
       调用被调用方函数。

       参数：
           - **payload** - 请求参数，需为 JSON 字符串格式。
       
       返回：
           响应结果，类型为 ObjectRef。
        
       抛出：
           :cpp:class:`FunctionError` - 执行请求时抛出的异常。
    
    .. cpp:function:: Function &Options(const InvokeOptions &opt)
 
       设置调用选项。

       参数：
           - **opt** - 参考 InvokeOptions 类型。
       
       返回：
           InvokeOptions 类型的引用。
        
    .. cpp:function:: const std::string GetObjectRef(ObjectRef &objectRef)

       设置调用选项。

       参数：
           - **objectRef** - ObjectRef 类型的引用。
       
       返回：
           调用结果。

    .. cpp:function:: const std::shared_ptr<Context> GetContext() const

       获取已配置的 Context 参数。

       参数：
           - **objectRef** - ObjectRef 类型的引用。
       
       返回：
           已配置的 Context。

       抛出：
           :cpp:class:`FunctionError` - 获取调用结果时抛出的异常。

.. cpp:class:: ObjectRef

    ObjectRef 类。

    **公共函数**
 
    .. cpp:function:: const std::string Get()
 
       获取 objectRef 的结果。

       返回：
           objectRef 的结果。

.. cpp:class:: FunctionError : public std::exception

    异常类。

    .. note::
        异常类错误码可以参考最下方的 ErrorCode 章节。

    **公共函数**
 
    .. cpp:function:: inline FunctionError(int code, const std::string message)
 
       FunctionError 的构造函数（用于初始化错误代码和消息）。

       参数：
           - **code** - 整数类型的错误代码（映射至 ErrorCode 枚举）。
           - **message** - 描述性错误消息。
        
       .. code-block:: cpp

          // 需要添加 #include "FunctionError.h"
          Function::FunctionLogger logger = context.GetLogger();
          try {
              auto func = Function::Function(context, "hello");
              auto ref = func.Invoke("{}");
              string result = ref.Get();
          } catch (Function::FunctionError e) {
              if (e.GetErrorCode() == Function::ErrorCode::FUNCTION_EXCEPTION) {  // need add #include "Constant.h"
                  logger.Error("err is FUNCTION_EXCEPTION, errCode is %d, errMsg is %s",
                      e.GetErrorCode(), e.GetMessage());
              } else {
                  logger.Error("errCode is %d, errMsg is %s", e.GetErrorCode(), e.GetMessage());
              }
          }

    .. cpp:function:: virtual ~FunctionError() = default
 
       默认析构函数。
    
    .. cpp:function:: const char *what() const noexcept override
 
       获取异常消息的 char* 类型输出。

       返回：
           异常消息。
    
    .. cpp:function:: ErrorCode GetErrorCode() const
 
       获取异常码。

       返回：
           异常码。

    .. cpp:function:: const std::string GetMessage() const
 
       获取异常消息。

       返回：
           异常消息。
    
    .. cpp:function:: const std::string GetJsonString() const
 
       获取异常信息并以 JSON 格式显示。

       返回：
           异常信息。

.. cpp:enum:: Function::ErrorCode

    用于运行时和用户处理程序操作的错误码。

    .. cpp:enumerator:: OK = 0

        成功
    
    .. cpp:enumerator:: ERROR = 1

        通用错误
    
    .. cpp:enumerator:: ILLEGAL_ACCESS = 4001

        非法访问运行时

    .. cpp:enumerator:: FUNCTION_EXCEPTION = 4002

        用户函数异常
    
        .. cpp:enumerator:: ILLEGAL_RETURN = 4004

        用户处理程序返回了非法值

    .. cpp:enumerator:: USER_STATE_UNDEFINED_ERROR = 4005

        用户状态未定义错误

    .. cpp:enumerator:: USER_INITIALIZATION_FUNCTION_EXCEPTION = 4009

        用户初始化函数异常

    .. cpp:enumerator:: USER_LOAD_FUNCTION_EXCEPTION = 4014

        用户加载函数异常

    .. cpp:enumerator:: NO_SUCH_INSTANCE_NAME_ERROR_CODE = 4026

        实例名称未找到

    .. cpp:enumerator:: INVALID_PARAMETER = 4040

        无效参数

    .. cpp:enumerator:: NO_SUCH_STATE_ERROR_CODE = 4041

        非法访问运行时（状态不存在）

    .. cpp:enumerator:: INTERNAL_ERROR = 110500

        内部系统错误
    
