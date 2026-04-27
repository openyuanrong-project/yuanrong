/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once
#include <openssl/ssl.h>
#include <openssl/x509v3.h>

#include <string>

#include "sensitive_data.h"

namespace YR {
namespace Libruntime {
SensitiveData GetPrivateKey(EVP_PKEY *pkey);
std::string GetCert(X509 *cert);
std::string GetCa(const stack_st_X509 *ca);
EVP_PKEY *GetPrivateKeyFromFile(const std::string &keyFile, char *password);
X509 *GetCertFromFile(const std::string &certFile);
STACK_OF(X509) * GetCAFromFile(const std::string &caFile);
void ClearPemCerts(EVP_PKEY *pkey, X509 *cert, STACK_OF(X509) * ca);
}  // namespace Libruntime
}  // namespace YR
