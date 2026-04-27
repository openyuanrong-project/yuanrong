# wait

package: `org.yuanrong.api`.

## Interface description

### public static WaitResult wait(List<ObjectRef> refs, int waitNum, int timeoutSec) throws YRException

Wait for the results to return.

This method returns when the number of returned results reaches waitNum or the wait time exceeds timeoutSec.

```java

int y = 1;
// Get the values of multiple object refs in parallel.
List<ObjectRef> objectRefs = new ArrayList<>();
for (int i = 0; i < 3; i++) {
    objectRefs.add(YR.put(i));
}
WaitResult waitResult = YR.wait(objectRefs, /*num_returns=*/ 1, /*timeoutMs=*/ 1000);
System.out.println(waitResult.getReady()); // List of ready objects.
System.out.println(waitResult.getUnready()); // list of unready objects.
```

- Parameters:

   - **refs** – ObjectRef list.
   - **waitNum** - The number of results to be returned. The value must be greater than 0.
   - **timeoutSec** – The timeout time for waiting, in seconds. The value must be greater than or equal to 0 or equal to -1.

- Returns:

    WaitResult: Store the returned results. Use getReady() to get a list of ObjectRef that can be retrieved, and use getUnready() to get a list of ObjectRef that cannot be retrieved.

- Throws:

   - **YRException** - Unified exception types thrown.
