# Collective

## CollectiveGroup

The parameter structure is supplemented with the following explanation:

```{doxygenstruct} YR::Collective::CollectiveGroupSpec
    :members:
```

Type and enumeration definitions:

```{doxygenenum} YR::Collective::Backend
```

```{doxygenenum} YR::DataType
```

```{doxygenenum} YR::ReduceOp
```

Constants:

```{doxygenvariable} YR::Collective::DEFAULT_COLLECTIVE_TIMEOUT
```

```{doxygenvariable} YR::Collective::DEFAULT_GROUP_NAME
```

## Collective-GroupOps

Collective communication group management interfaces for creating, initializing, and destroying collective communication groups.

```{doxygenfunction} YR::Collective::InitCollectiveGroup(const CollectiveGroupSpec &groupSpec, int rank, const std::string &prefix = "")
```

```{doxygenfunction} YR::Collective::CreateCollectiveGroup(const CollectiveGroupSpec &groupSpec, const std::vector<std::string> &instanceIDs, const std::vector<int> &ranks)
```

```{doxygenfunction} YR::Collective::DestroyCollectiveGroup(const std::string &groupName)
```

```{doxygenfunction} YR::Collective::GetWorldSize(const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::GetRank(const std::string &groupName = DEFAULT_GROUP_NAME)
```

## Collective-CommOps

Collective communication operation interfaces, providing distributed communication primitives such as AllReduce, Reduce, AllGather, Broadcast, etc.

```{doxygenfunction} YR::Collective::AllReduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Reduce(const void *sendbuf, void *recvbuf, int count, DataType dtype, const ReduceOp &op, int dstRank, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::AllGather(const void *sendbuf, void *recvbuf, int count, DataType dtype, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Broadcast(const void *sendbuf, void *recvbuf, int count, DataType dtype, int srcRank, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Scatter(const std::vector<void *> sendbuf, void *recvbuf, int count, DataType dtype, int srcRank, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Barrier(const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Send(const void *sendbuf, int count, DataType dtype, int dstRank, int tag = 0, const std::string &groupName = DEFAULT_GROUP_NAME)
```

```{doxygenfunction} YR::Collective::Recv(void *recvbuf, int count, DataType dtype, int srcRank, int tag = 0, const std::string &groupName = DEFAULT_GROUP_NAME)
```
