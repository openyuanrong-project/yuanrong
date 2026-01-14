# Context

package:`com.services.runtime`

## public interface Context

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

#### int getRemainingTimeInMilliSeconds()

Get the remaining time from the current time to the timeout time.

- Returns:

    remainingTime (int): remaining time from the current time to the timeout time.

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

#### int getRunningTimeInSeconds()

Get the time distributed to the running of the function, when exceed the specified time, the running of the function would be stopped by force.

- Returns:
 
    runningTime (int): The time distributed to the running of the function.

#### int getMemorySize()

Get the memory size distributed the running function.

- Returns:
  
    memorySize (int): Function memory size, unit: MB.

#### int getCPUNumber()

Get the cpu size distributed the running function.

- Returns:
 
    cpuNumber (int): Function CPU number, unit: m, 1C=1000m.

#### String getAlias()

Get function alias.

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
