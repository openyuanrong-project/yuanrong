# get

package: `org.yuanrong.api`.

## Interface description

### public static Object get(ObjectRef ref, int timeoutSec) throws YRException

Get the value stored in the data system.

This method will block for timeoutSec seconds until the result is obtained. If the result is not obtained within the timeout period, an exception is thrown. The timeout period should be greater than 0, otherwise, an YRException(“timeout is invalid, it needs greater than 0”) will be thrown immediately.

```java

ObjectRef ref = YR.put(1);
System.out.println(YR.get(ref,3000));
```

- Parameters:

   - **ref** – invoke method or YR.put return value, required.
   - **timeoutSec** – Timeout time for obtaining data in seconds, required. The value must be greater than 0.

- Returns:

    Object The obtained object.

- Throws:

   - **YRException** - Error retrieving data, timeout or data system service unavailable.
