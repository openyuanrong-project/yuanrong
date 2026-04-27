# KV().MSetTx

```{doxygenfunction} YR::KVManager::MSetTx(const std::vector<std::string> &keys, const std::vector<char *> &vals, const std::vector<size_t> &lens, ExistenceOpt existence)
```

```{doxygenfunction} YR::KVManager::MSetTx(const std::vector<std::string> &keys, const std::vector<std::string> &vals, ExistenceOpt existence)
```

```{doxygenfunction} YR::KVManager::MSetTx(const std::vector<std::string> &keys, const std::vector<char *> &vals, const std::vector<size_t> &lens, const MSetParam &mSetParam)
```

```{doxygenfunction} YR::KVManager::MSetTx(const std::vector<std::string> &keys, const std::vector<std::string> &vals, const MSetParam &mSetParam)
```

The parameter structure is supplemented with the following explanation:

```{doxygenstruct} YR::MSetParam
    :members:
```
