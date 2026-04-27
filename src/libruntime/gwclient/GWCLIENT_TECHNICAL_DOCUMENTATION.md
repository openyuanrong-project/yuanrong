# Gateway Client (gwclient) Technical Documentation

## Overview

The `gwclient` is a gateway client library in the openYuanrong runtime that provides a unified interface for serverless functions to communicate with backend services through HTTP/HTTPS protocols. It implements multiple storage abstractions including Object Store, State Store, Stream Store, and Hetero Store.

## Architecture

### Component Structure

```
gwclient/
├── gw_client.h/cpp           # Main client implementation
├── gw_datasystem_client_wrapper.h  # Datasystem wrapper
└── http/
    ├── http_client.h         # Base HTTP client interface
    ├── client_manager.h/cpp  # Connection pool management
    ├── async_http_client.h/cpp   # Plain HTTP client
    └── async_https_client.h/cpp  # HTTPS client (TLS/mTLS)
```

### Class Hierarchy

```
HttpClient (abstract)
├── AsyncHttpClient      # Plain HTTP (no encryption)
└── AsyncHttpsClient     # HTTPS with TLS/mTLS

ClientManager : HttpClient  # Connection pool manager

GwClient : FSIntf, ObjectStore, StateStore, StreamStore, HeteroStore
└── ClientBuffer : NativeBuffer  # Client-side buffer wrapper
```

---

## Authentication

### Header-Based Authentication

The client builds authentication headers through `BuildHeaders()` method:

```cpp
std::unordered_map<std::string, std::string> GwClient::BuildHeaders(
    const std::string &remoteClientId,
    const std::string &traceId,
    const std::string &tenantId)
```

**Standard Headers:**
| Header Key | Description |
|------------|-------------|
| `remoteClientId` / `X-Remote-Client-Id` | Job/Client identifier |
| `traceId` / `X-Trace-Id` | Request tracing ID |
| `tenantId` / `X-Tenant-Id` | Tenant identifier |
| `X-Auth` | Authentication token |

### AK/SK Signature Authentication

For advanced authentication, the client supports AWS-style AK/SK signing:

```cpp
std::pair<std::unordered_map<std::string, std::string>, std::string> 
GwClient::BuildRequestWithAkSk(const std::shared_ptr<InvokeSpec> spec, const std::string &url)
```

The signing is performed by `SignHttpRequest()` which generates cryptographic request signatures.

### TLS/mTLS Support

The `ClientManager` supports three security modes:

1. **No TLS**: Plain HTTP connections
2. **One-way TLS**: Server certificate verification only
3. **mTLS (Mutual TLS)**: Both client and server certificates verified

```cpp
// TLS Configuration in ClientManager::InitCtxAndIocThread()
if (enableMTLS) {
    // Load client certificate for mutual authentication
    ctx->use_certificate_chain_file(librtCfg->certificateFilePath);
    ctx->use_private_key_file(librtCfg->privateKeyPath, ssl::context::pem);
    // Verify server certificate
    ctx->set_verify_mode(ssl::verify_peer);
} else if (enableTLS_) {
    // One-way TLS: verify server only
    ctx->set_verify_mode(ssl::verify_peer);
}
```

---

## Lease Management

### Lifecycle

The `GwClient` manages lease lifecycle through three operations:

1. **Lease Acquisition** (`Lease()`): Initial lease request
2. **Lease Keepalive** (`KeepLease()`): Periodic keepalive
3. **Lease Release** (`Release()`): Explicit release

### API Endpoints

| Operation | HTTP Method | Endpoint |
|-----------|-------------|----------|
| Acquire | PUT | `/client/v1/lease` |
| Keepalive | POST | `/client/v1/lease/keepalive` |
| Release | DELETE | `/client/v1/lease` |

### Implementation

```cpp
ErrorInfo GwClient::Start(...) {
    auto error = Lease();  // Acquire lease on start
    timer_ = timerWorker_->CreateTimer(KEEPALIVE_INTERVAL, KEEPALIVE_TIMES, [this](void) {
        this->KeepLease();  // Periodic keepalive
    });
}

ErrorInfo GwClient::HandleLease(const std::string &url, const http::verb &verb) {
    auto req = BuildLeaseRequest();
    req.set_remoteclientid(this->jobId_);
    httpClient_->SubmitInvokeRequest(verb, url, headers, body, ...);
}
```

**Configuration:**
- `KEEPALIVE_INTERVAL`: 60 seconds (60 * 1000 ms)
- `KEEPALIVE_TIMES`: Unlimited (-1)

### Failure Handling

The client tracks lease failures and logs warnings:

```cpp
if (!error.OK()) {
    YRLOG_WARN("Keepalive fails for {} consecutive times.", ++lostLeaseTimes_);
} else {
    lostLeaseTimes_ = 0;
}
```

---

## Storage Operations

### Object Storage (POSIX Semantics)

The client provides object storage operations through the ObjectStore interface.

#### Put Operation

```cpp
ErrorInfo GwClient::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                       const std::unordered_set<std::string> &nestedID, 
                       const CreateParam &createParam);
```

**Endpoint:** `POST /datasystem/v1/obj/put`

**Request Build:**
```cpp
PutRequest GwClient::BuildObjPutRequest(...) {
    req.set_objectid(objID);
    req.set_objectdata(data->ImmutableData(), data->GetSize());
    req.set_writemode(static_cast<int32_t>(createParam.writeMode));
    req.set_consistencytype(static_cast<int32_t>(createParam.consistencyType));
    req.set_cachetype(static_cast<int32_t>(createParam.cacheType));
}
```

#### Get Operation

```cpp
SingleResult GwClient::Get(const std::string &objID, int timeoutMS);
MultipleResult GwClient::Get(const std::vector<std::string> &ids, int timeoutMS);
```

**Endpoint:** `POST /datasystem/v1/obj/get`

#### Reference Management

```cpp
ErrorInfo GwClient::IncreGlobalReference(const std::vector<std::string> &objectIds);
ErrorInfo GwClient::DecreGlobalReference(const std::vector<std::string> &objectIds);
```

**Endpoints:**
- Increase: `POST /datasystem/v1/obj/increaseref`
- Decrease: `POST /datasystem/v1/obj/decreaseref`

**Implementation Note:** Decrease operations use async processing through `AsyncDecreRef` for better performance.

### Key-Value Storage

The client implements StateStore interface for KV operations.

#### Set Operation

```cpp
ErrorInfo GwClient::Write(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam);
```

**Endpoint:** `POST /datasystem/v1/kv/set`

#### Get Operation

```cpp
SingleReadResult GwClient::Read(const std::string &key, int timeoutMS);
MultipleReadResult GwClient::Read(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial);
```

**Endpoint:** `POST /datasystem/v1/kv/get`

#### Multi-Key Transaction

```cpp
ErrorInfo GwClient::MSetTx(const std::vector<std::string> &keys, 
                          const std::vector<std::shared_ptr<Buffer>> &vals,
                          const MSetParam &mSetParam);
```

**Endpoint:** `POST /datasystem/v1/kv/msettx`

#### Delete Operation

```cpp
ErrorInfo GwClient::Del(const std::string &key);
MultipleDelResult GwClient::Del(const std::vector<std::string> &keys);
```

**Endpoint:** `POST /datasystem/v1/kv/del`

### Buffer Management

The `ClientBuffer` class provides a wrapper for client-side buffers:

```cpp
class ClientBuffer : public NativeBuffer {
    virtual ErrorInfo Seal(const std::unordered_set<std::string> &nestedIds) override;
};
```

When sealed, the buffer data is automatically uploaded via `PosixObjPut()`.

---

## HTTP Communication Layer

### Connection Pool

The `ClientManager` maintains a pool of HTTP clients:

```cpp
class ClientManager : public HttpClient {
    std::vector<std::shared_ptr<HttpClient>> clients;  // Connection pool
    uint32_t connectedClientsCnt_;                      // Active connections
    uint32_t maxConnSize_;                             // Max pool size
};
```

**Features:**
- Configurable pool size via `maxConnSize_`
- Thread-safe connection allocation
- Automatic reconnection on connection failure

### Request Flow

```
ClientManager::SubmitInvokeRequest()
    │
    ▼
Select available client from pool
    │
    ▼
Check connection活性 (IsConnActive)
    │
    ├── Active ──► Submit request
    │
    └── Inactive ──► ReInit() ──► Submit request
```

### Retry Mechanism

The HTTP client implements automatic retry on failure:

```cpp
void AsyncHttpClient::OnWrite(...) {
    if (!retried_) {
        retried_ = true;
        if (ReInit(requestId).OK()) {
            // Retry request
        }
    }
}
```

### Connection Lifecycle

```cpp
void AsyncHttpClient::GracefulExit() noexcept {
    stream_.socket().shutdown(asio::ip::tcp::socket::shutdown_both, ec);
    stream_.close();
}
```

---

## Function Invocation

### Async Invocation

The client supports function invocation through the gateway:

```cpp
void GwClient::InvocationAsync(const std::string &url, 
                              const std::shared_ptr<InvokeSpec> spec,
                              const InvocationCallback &callback);
```

### Instance Lifecycle Operations

```cpp
void GwClient::CreateAsync(...);    // Create function instance
void GwClient::InvokeAsync(...);    // Invoke function
void GwClient::KillAsync(...);     // Kill function instance
```

**Endpoints:**
| Operation | Endpoint |
|-----------|----------|
| Create | `/serverless/v1/posix/instance/create` |
| Invoke | `/serverless/v1/posix/instance/invoke` |
| Kill | `/serverless/v1/posix/instance/kill` |

---

## Error Handling

### Response Code Mapping

```cpp
bool IsResponseSuccessful(const uint statusCode) {
    return (statusCode >= 200 && statusCode <= 299);
}

bool IsResponseServerError(const uint statusCode) {
    return (statusCode >= 500 && statusCode <= 599);
}
```

### Error Generation

```cpp
ErrorInfo GwClient::GenRspError(const boost::beast::error_code &errorCode, 
                               const uint statusCode, 
                               const std::string &result,
                               const std::shared_ptr<std::string> requestId) {
    if (errorCode) {
        return ErrorInfo(ERR_INNER_COMMUNICATION, ss.str());
    }
    if (!IsResponseSuccessful(statusCode)) {
        return ErrorInfo(ERR_PARAM_INVALID, ss.str());
    }
    return ErrorInfo::OK();
}
```

---

## Configuration

### Connection Parameters

```cpp
struct ConnectionParam {
    std::string ip;
    std::string port;
    int idleTime{120};        // Connection idle timeout (seconds)
    int timeoutSec = CONNECTION_NO_TIMEOUT;
};
```

### TLS Configuration

| Parameter | Description |
|-----------|-------------|
| `enableMTLS` | Enable mutual TLS |
| `enableTLS_` | Enable one-way TLS |
| `certificateFilePath` | Client certificate (mTLS) |
| `privateKeyPath` | Client private key (mTLS) |
| `verifyFilePath` | CA certificate for server verification |
| `skipServerVerify` | Skip server certificate verification |

---

## Thread Safety

### Thread-Local Storage

```cpp
thread_local static std::string threadLocalTenantId;
```

The tenant ID is stored in thread-local storage for per-thread isolation.

### Mutex Protection

Connection pool uses `absl::Mutex` for thread-safe access:

```cpp
mutable absl::Mutex mu_;
bool isUsed_ ABSL_GUARDED_BY(mu_);
bool isConnectionAlive_ ABSL_GUARDED_BY(mu_);
```

---

## Dependencies

- **Boost.Asio**: Async I/O operations
- **Boost.Beast**: HTTP/HTTPS protocol
- **Boost.Asio SSL**: TLS/SSL support
- **nlohmann/json**: JSON parsing
- **Protocol Buffers**: Serialization

---

## Usage Example

```cpp
// Create client
auto gwClient = std::make_shared<GwClient>(funcId, handlers, security);

// Initialize with HTTP client
auto clientManager = std::make_shared<ClientManager>(librtCfg);
clientManager->Init({ip, port});
gwClient->Init(clientManager);

// Start and acquire lease
gwClient->Start(jobId, instanceId, runtimeId, functionName);

// Storage operations
gwClient->Put(buffer, objectId, nestedIds, createParam);
auto [err, result] = gwClient->Get(objectId, timeoutMS);

// Cleanup
gwClient->Stop();
```
