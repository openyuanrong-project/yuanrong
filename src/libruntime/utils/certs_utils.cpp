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

#include "certs_utils.h"

#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
using namespace YR::utility;
SensitiveData GetPrivateKey(EVP_PKEY *pkey)
{
    if (!pkey) {
        YRLOG_WARN("failed to get pkey from EVP_PKEY, empty pkey");
        return {};
    }

    BIO *mem = BIO_new(BIO_s_mem());
    (void)PEM_write_bio_PrivateKey(mem, pkey, nullptr, nullptr, 0, nullptr, nullptr);
    char *data;
    auto len = BIO_get_mem_data(mem, &data);
    auto result = SensitiveData(data, len);
    // clear
    if (int ret = memset_s(data, len, 0, len); ret != 0) {
        YRLOG_WARN("data memset 0 failed, ret: {}", ret);
    }
    (void)BIO_free(mem);
    return result;
}

std::string GetCert(X509 *cert)
{
    if (!cert) {
        YRLOG_WARN("failed to get cert from X509, empty cert");
        return "";
    }

    BIO *mem = BIO_new(BIO_s_mem());
    (void)PEM_write_bio_X509(mem, cert);
    char *data;
    long len = BIO_get_mem_data(mem, &data);
    if (len <= 0) {
        YRLOG_WARN("BIO get invalid data length");
        return "";
    }
    std::string result = std::string(data, len);
    (void)BIO_free(mem);
    return result;
}

std::string GetCa(const stack_st_X509 *ca)
{
    if (!ca || !sk_X509_num(ca)) {
        YRLOG_WARN("failed to get ca from stack_st_X509, empty ca");
        return "";
    }

    std::string result = "";
    for (int i = 0; i < sk_X509_num(ca); i++) {
        X509 *x = sk_X509_value(ca, i);
        BIO *mem = BIO_new(BIO_s_mem());
        (void)PEM_write_bio_X509(mem, x);
        char *data;
        long len = BIO_get_mem_data(mem, &data);
        if (len <= 0) {
            return "";
        }
        result += std::string(data, len);
        (void)BIO_free(mem);
    }
    return result;
}

EVP_PKEY *GetPrivateKeyFromFile(const std::string &keyFile, char *password)
{
    FILE *fp = fopen(keyFile.c_str(), "r");
    if (!fp) {
        YRLOG_ERROR("unable to open key {}", keyFile);
        return nullptr;
    }

    EVP_PKEY *key = PEM_read_PrivateKey(fp, nullptr, nullptr, password);
    if (!key) {
        YRLOG_ERROR("unable to parse key in {}", keyFile);
        fclose(fp);
        return nullptr;
    }
    fclose(fp);
    return key;
}

X509 *GetCertFromFile(const std::string &certFile)
{
    if (!IsAbsolute(certFile)) {
        YRLOG_ERROR("invalid cert file path {}", certFile);
        return nullptr;
    }
    FILE *fp = fopen(certFile.c_str(), "r");
    if (!fp) {
        YRLOG_ERROR("unable to open cert {}", certFile);
        return nullptr;
    }

    X509 *cert = PEM_read_X509(fp, nullptr, nullptr, nullptr);
    if (!cert) {
        YRLOG_ERROR("unable to parse certificate in {}", certFile);
        (void)fclose(fp);
        return nullptr;
    }
    (void)fclose(fp);
    return cert;
}

STACK_OF(X509) * GetCAFromFile(const std::string &caFile)
{
    FILE *caCertFile = fopen(caFile.c_str(), "r");
    if (!caCertFile) {
        YRLOG_ERROR("Failed to open CA certificate file: {}", caFile);
        return nullptr;
    }

    STACK_OF(X509_INFO) *caCertInfoStack = PEM_X509_INFO_read(caCertFile, nullptr, nullptr, nullptr);
    fclose(caCertFile);

    if (!caCertInfoStack) {
        YRLOG_ERROR("Failed to read CA certificate information from file: {}", caFile);
        return nullptr;
    }

    STACK_OF(X509) *caCertStack = sk_X509_new_null();
    for (int i = 0; i < sk_X509_INFO_num(caCertInfoStack); i++) {
        X509_INFO *certInfo = sk_X509_INFO_value(caCertInfoStack, i);
        if (certInfo->x509) {
            sk_X509_push(caCertStack, X509_dup(certInfo->x509));
        }
    }

    sk_X509_INFO_pop_free(caCertInfoStack, X509_INFO_free);

    if (sk_X509_num(caCertStack) == 0) {
        YRLOG_ERROR("No CA certificates found in file: {}", caFile);
        sk_X509_free(caCertStack);
        return nullptr;
    }

    return caCertStack;
}

void ClearPemCerts(EVP_PKEY *pkey, X509 *cert, STACK_OF(X509) * ca)
{
    if (pkey) {
        EVP_PKEY_free(pkey);
    }
    if (cert) {
        X509_free(cert);
    }
    if (ca) {
        sk_X509_pop_free(ca, X509_free);
    }
}
}  // namespace Libruntime
}  // namespace YR
