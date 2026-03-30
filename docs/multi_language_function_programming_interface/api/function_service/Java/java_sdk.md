# Java SDK

## public interface Context

package:`org.yuanrong.services.runtime`

Provide function runtime capability to user.

```java

public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        RuntimeLogger logger = context.getLogger();
        logger.Log("test");
        return "";
    }
}
```

### Interface description

#### String getRequestID()

Get trace ID of one request.

- Returns:
  
    requestID (String): Can be used to uniquely identify a request, usually used for log to trace the chain of a request.

#### String getUserData(String key)

The method to get environment and encrypt env.

- Parameters:

   - **key** (Sring) - Env's key.
  
- Returns:
   
    value (Sring): Env value or encrypt env value.

#### String getFunctionName()

Get function name.

- Returns:
  
    functionName (Sring): The name of function self

#### int getMemorySize()

Get the memory size distributed the running function.

- Returns:
  
    memorySize (int): Function memory size, unit: MB.

#### int getCPUNumber()

Get the cpu size distributed the running function.

- Returns:
 
    cpuNumber (int): Function CPU number, unit: m, 1C=1000m.

#### String getInstanceLabel()

Get the instance label.

- Returns:

    instanceLabel (String): The instance label.

#### RuntimeLogger getLogger()

Gets the logger for user to log out in standard output, The Logger interface must be provided in SDK.

- Returns:
  
    logger (RuntimeLogger): Runtime logger, use to print log.

```java

public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        RuntimeLogger logger = context.getLogger();
        logger.log("test log"):
        return"";
    }
}
```

#### String getSessionId()

Get the agent session ID associated with the current invocation.

Returns an empty string when the request does not carry a session ID or `use_agent_session=false`.

- Returns:

    sessionId (String): Session ID.

#### SessionService getSessionService()

Get the session service for the current invocation.

Returns `null` if no session is associated with the current request.

- Returns:

    SessionService: Session service instance, see SessionService interface definition below.

## public class Function

package:`org.yuanrong.function`

Function class are used for invocation between functions by function name.

### Interface description

#### public Function(Context context, String functionNameWithVersion)

Init Function object with user infomation by context and target invocation function name and version.

- Parameters:

   - **context** (Context) - Init user's tenant, service and more info to Function object, always set with user function  input parameter, if not specidied will use current function default infomation.
  
   - **functionNameWithVersion** (String) - Designated target invocation function and version, it will throw error when invoke without this parameter.

#### public Function(String functionNameWithVersion)

Init Function object with functionNameWithVersion, context will set to null.

#### public Function(Context context)

Init Function object with context, functionNameWithVersion will set to null.

```java

public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        String funcName = jsonObject.get("func_name").getAsString();
        Function func = new Function(context, funcName);
        return "";
    }
}
```

#### public Function options(CreateOptions opts)

Configuring target function instance include cpu and memory, if doesn't equal target function current config will start a new instance.

- Parameters:

   - opts:
  
     - **cpu** (int) - Target function instance cpu, unit: 1m, 1C=1000m.
    
     - **memory** (int) - Target function instance memory, unit: MB.

#### public <T> ObjectRef<T> invoke(String payload)

The method to invoke other function, this is a async method.

- Parameters:
  
   - **payload** (String) - Invoke body, should be json type.
  
- Returns:
  
    ObjectRef (ObjectRef<T>): The result of inoke other function.
  
- Throws:
  
   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

```java
  
public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        String funcName = jsonObject.get("func_name").getAsString();
        Function func = new Function(context, funcName);
        ObjectRef<String> obj = func.invoke("{}")
        return obj.get(String.class);
    }
}
```

# InvokeException

package:`org.yuanrong.function.runtime.exception`

## public class InvokeException extends RuntimeException

The error type yuanrong sdk invoke return.

### Interface description

#### public int getErrorCode()

Get error's code.

- Returns:

    errorCode (int): Invoke error code.

#### public String getMessage()

Get error message.

- Returns:

    message (String): Error message.

#### public String toString()

Assemble error code and message into a fixed format.

- Returns:
   
    errorMsg (String): `{\"code\":\"errorCode\", \"message\":\"message\"}`

# CreateOptions

package:`org.yuanrong.function`

## Public class CreateOptions

Use to invoke.

- Parameters:
  
   - **cpu** (int) - Invoke target instance cpu, unit: m, 1C=1000m.
  
   - **memory** (int) - Invoke target instance memory, unit: MB.
  
   - **aliasParams** (Map<String, String>) - Target function alias.

### Interface description

#### public CreateOptions(int cpu, int memory, Map<String, String> aliasParams)

Init create option object.

- Parameters:
  
   - **cpu** (int) - Invoke target instance cpu, unit: m, 1C=1000m.
  
   - **memory** (int) - Invoke target instance memory, unit: MB.
  
   - **aliasParams** (Map<String, String>) - Target function alias.

#### public CreateOptions(int cpu, int memory)

Init without aliasParams.

#### public CreateOptions(Map<String, String> aliasParams)

Init with 0 cpu and 0 memory.

#### public CreateOptions(int memory)

Init with 0 cpu and empty aliasParams.
  
# ObjectRef<T>

package:`org.yuanrong.function`

## public class ObjectRef<T>

Object used to recive invoke return.

### Interface description

#### public T get(int timeoutSec)

Get function.invoke() result with time limit, the mothod will set result to object of ObjectRef and return result, if the time taken to recive the return exceeds the limit of timeoutSec will throw error.

- Parameters:
  
   - **timeoutSec** (int) - The time limit of recive result.
  
- Returns:
  
    T (T): Result of converting to T-shape.
  
- Throws:
  
   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

#### public T get()

Call get(int timeoutSec) without timeout.

#### public T get(Class<?> classType, int timeoutSec)

Besides recive and store result, this method will check if result complies with classType format and transfer result to classType.

- Parameters:

   - **timeoutSec** (int) - The time limit of recive result.

   - **classType** (Class<?>) - Should equal to T type.

- Returns:
 
    T: Result of converting to T-shape

- Throws:

   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

#### public T get(Class<?> classType)

Call get(Class<?> classType, int timeoutSec)get(int timeoutSec) without timeout.

# SessionService

package:`org.yuanrong.services.session`

SDK interface providing session access capability.

## public interface SessionService

Used to load the session object associated with the current invocation.

### Interface description

#### SessionObj loadSession()

Load the session associated with the current invocation.

Returns `null` when no `sessionId` is present in the request or `use_agent_session=false`.

- Returns:

    SessionObj: The current session object, see SessionObj interface definition below. Returns `null` if no session is associated with the current request.

- Throws:

   - **YRException** (YRException) - Thrown when the underlying JNI call fails.

```java

public class demo {
    public String handleRequest(JsonObject jsonObject, Context context) {
        SessionService sessionService = context.getSessionService();
        if (sessionService == null) {
            return "no session";
        }
        SessionObj session = sessionService.loadSession();
        if (session == null) {
            return "session not found";
        }
        return session.getID();
    }
}
```

# SessionObj

package:`org.yuanrong.services.session`

## public interface SessionObj

Represents an agent session object.

The session is managed by the runtime. Users should only read and modify the session via the provided accessors. Do NOT modify the list returned by `getHistories()` directly; always use `setHistories(List)` to write back changes so the runtime can be notified.

### Interface description

#### String getID()

Get the session ID.

- Returns:

    sessionID (String): Session ID.

#### List<String> getHistories()

Get a read-only snapshot of the conversation history.

The returned list is an unmodifiable view. Mutating it has no effect on the runtime state.

- Returns:

    histories (List<String>): Unmodifiable list of conversation history entries.

#### void setHistories(List<String> histories)

Update the conversation history.

This method immediately synchronizes the new value to libruntime, ensuring the runtime holds the latest state before persisting.

- Parameters:

   - **histories** (List<String>) - New history list (null treated as empty list).

- Throws:

   - **YRException** (YRException) - Thrown when the JNI call fails.

```java

public class demo {
    public String handleRequest(JsonObject jsonObject, Context context) {
        SessionService sessionService = context.getSessionService();
        if (sessionService == null) {
            return "no session";
        }
        SessionObj session = sessionService.loadSession();
        if (session == null) {
            return "session not found";
        }
        List<String> histories = new ArrayList<>(session.getHistories());
        histories.add("new message");
        session.setHistories(histories);
        return "history updated";
    }
}
```

# ManagedSessionObj

package:`org.yuanrong.services.session`

## public class ManagedSessionObj implements SessionObj

Runtime-managed implementation of SessionObj. Constructed by the JNI layer after loading from libruntime.

### Interface description

#### public ManagedSessionObj(String id, List<String> histories)

Constructor.

- Parameters:

   - **id** (String) - Session ID.
   - **histories** (List<String>) - Initial history list (may be empty, null treated as empty list).

#### public static ManagedSessionObj fromJson(String json)

Deserialize from the canonical JSON format returned by libruntime.

JSON format: `{"sessionID":"s-123","histories":["user: hello","assistant: hi"]}`

- Parameters:

   - **json** (String) - Session JSON string (may be null or empty).

- Returns:

    ManagedSessionObj: Deserialized object. Returns empty object if json is null, empty, or parsing fails.

# Function

package:`org.yuanrong.function`

## public class Function

Function class are used for invocation between functions by function name.

### Interface description

#### public Function(Context context, String functionNameWithVersion)

Init Function object with user infomation by context and target invocation function name and version.

- Parameters:

   - **context** (Context) - Init user's tenant, service and more info to Function object, always set with user function  input parameter, if not specidied will use current function default infomation.
  
   - **functionNameWithVersion** (String) - Designated target invocation function and version, it will throw error when invoke without this parameter.

#### public Function(String functionNameWithVersion)

Init Function object with functionNameWithVersion, context will set to null.

#### public Function(Context context)

Init Function object with context, functionNameWithVersion will set to null.

```java

public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        String funcName = jsonObject.get("func_name").getAsString();
        Function func = new Function(context, funcName);
        return "";
    }
}
```

#### public Function options(CreateOptions opts)

Configuring target function instance include cpu and memory, if doesn't equal target function current config will start a new instance.

- Parameters:

   - opts:
  
     - **cpu** (int) - Target function instance cpu, unit: 1m, 1C=1000m.
    
     - **memory** (int) - Target function instance memory, unit: MB.

#### public <T> ObjectRef<T> invoke(String payload)

The method to invoke other function, this is a async method.

- Parameters:
  
   - **payload** (String) - Invoke body, should be json type.
  
- Returns:
  
    ObjectRef (ObjectRef<T>): The result of inoke other function.
  
- Throws:
  
   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

```java
  
public class demo {
    public string handleRequest(JsonObject jsonObject, Context context) {
        String funcName = jsonObject.get("func_name").getAsString();
        Function func = new Function(context, funcName);
        ObjectRef<String> obj = func.invoke("{}")
        return obj.get(String.class);
    }
}
```

# InvokeException

package:`org.yuanrong.function.runtime.exception`

## public class InvokeException extends RuntimeException

The error type yuanrong sdk invoke return.

### Interface description

#### public int getErrorCode()

Get error's code.

- Returns:

    errorCode (int): Invoke error code.

#### public String getMessage()

Get error message.

- Returns:

    message (String): Error message.

#### public String toString()

Assemble error code and message into a fixed format.

- Returns:
   
    errorMsg (String): `{\"code\":\"errorCode\", \"message\":\"message\"}`

# CreateOptions

package:`org.yuanrong.function`

## Public class CreateOptions

Use to invoke.

- Parameters:
  
   - **cpu** (int) - Invoke target instance cpu, unit: m, 1C=1000m.
  
   - **memory** (int) - Invoke target instance memory, unit: MB.
  
   - **aliasParams** (Map<String, String>) - Target function alias.

### Interface description

#### public CreateOptions(int cpu, int memory, Map<String, String> aliasParams)

Init create option object.

- Parameters:
  
   - **cpu** (int) - Invoke target instance cpu, unit: m, 1C=1000m.
  
   - **memory** (int) - Invoke target instance memory, unit: MB.
  
   - **aliasParams** (Map<String, String>) - Target function alias.

#### public CreateOptions(int cpu, int memory)

Init without aliasParams.

#### public CreateOptions(Map<String, String> aliasParams)

Init with 0 cpu and 0 memory.

#### public CreateOptions(int memory)

Init with 0 cpu and empty aliasParams.
  
# ObjectRef<T>

package:`org.yuanrong.function`

## public class ObjectRef<T>

Object used to recive invoke return.

### Interface description

#### public T get(int timeoutSec)

Get function.invoke() result with time limit, the mothod will set result to object of ObjectRef and return result, if the time taken to recive the return exceeds the limit of timeoutSec will throw error.

- Parameters:
  
   - **timeoutSec** (int) - The time limit of recive result.
  
- Returns:
  
    T (T): Result of converting to T-shape.
  
- Throws:
  
   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

#### public T get()

Call get(int timeoutSec) without timeout.

#### public T get(Class<?> classType, int timeoutSec)

Besides recive and store result, this method will check if result complies with classType format and transfer result to classType.

- Parameters:

   - **timeoutSec** (int) - The time limit of recive result.

   - **classType** (Class<?>) - Should equal to T type.

- Returns:
 
    T: Result of converting to T-shape

- Throws:

   - **InvokeException** (InvokeException) - OpenYuanrong normal exception, specific meaning see InvokeException class.

#### public T get(Class<?> classType)

Call get(Class<?> classType, int timeoutSec)get(int timeoutSec) without timeout.
