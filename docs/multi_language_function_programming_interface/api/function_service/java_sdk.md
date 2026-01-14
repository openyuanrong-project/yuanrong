# Function

package:`com.function`

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

package:`com.function.runtime.exception`

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

package:`com.function`

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

package:`com.function`

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
