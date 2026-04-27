# ObjectRef

package: `org.yuanrong.runtime.client`.

## public class ObjectRef

ObjectRef class.

:::{Note}

1.openYuanrong encourages users to store large objects in the data system using YR.put and obtain a unique ObjectRef (object reference). When invoking user functions, use the ObjectRef instead of the object itself as a function parameter to reduce the overhead of transmitting large objects between openYuanrong and user function components, ensuring efficient flow.

2.The return value of each user function call will also be returned in the form of an ObjectRef, which the user can use as an input parameter for the next call or retrieve the corresponding object through YR.get.

3.Currently, users cannot construct objectRef by themselves.

:::

### Interface description

#### public ObjectRef(String objectID)

The constructor for ObjectRef.

- Parameters:

   - **objectID** - object ID in the openYuanrong cluster.

#### public ObjectRef(String objectID, Class<?> returnType)

The constructor for ObjectRef.

- Parameters:

   - **objectID** - object ID in the openYuanrong cluster.
   - **returnType** - object type.

#### public String getObjId()

Get objectID.

- Returns:

    objectID: object ID in the openYuanrong cluster.

#### public Object get(Class<?> classType) throws YRException, LibRuntimeException

Get results.

- Parameters:

   - **classType** - Class Type.

- Returns:

    The result of ObjectRef.

- Throws:

   - **YRException** - Unified exception types thrown.
   - **LibRuntimeException** - Data system error.

#### public Object get() throws YRException, LibRuntimeException

Get results.

- Returns:

    The result of ObjectRef.

- Throws:

   - **YRException** - Unified exception types thrown.
   - **LibRuntimeException** - Data system error.

#### public Object get(int timeout) throws YRException, LibRuntimeException

Get results.

- Parameters:

   - **timeout** - timeout duration in seconds.

- Returns:

    The result of ObjectRef.

- Throws:

   - **YRException** - Unified exception types thrown.
   - **LibRuntimeException** - Data system error.

### Private Members

``` java

private Class<?> type
```

object type

``` java

private boolean isByteBuffer = false
```

Whether the object is of type ByteBuffer.

``` java

private final String objectID
```

Object ID in the openYuanrong cluster.
