# Get

Retrieve the value of an object from the backend storage based on its key.
The interface call will block until the object's value is retrieved or a timeout occurs.

```{doxygenfunction} YR::Get(const ObjectRef<T> &, int)
```

```{doxygenfunction} YR::Get(const std::vector<ObjectRef<T>> &, int, bool)
```

```{doxygenvariable} YR::DEFAULT_GET_TIMEOUT_SEC
```
