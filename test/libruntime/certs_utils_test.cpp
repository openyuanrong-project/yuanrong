/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include <stdlib.h>
#include <fstream>

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "src/libruntime/utils/certs_utils.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::Libruntime;
using namespace YR::utility;

namespace YR {
namespace test {
class CertsUtilsTest : public ::testing::Test {
public:
    CertsUtilsTest()
    {
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-runtime",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
    }
    ~CertsUtilsTest() {}
};

static void GenPemCert(EVP_PKEY *pkey, X509 *cert, EVP_PKEY *caPkey = nullptr, X509 *caCert = nullptr)
{
    // generate a private key
    BIGNUM *bne = BN_new();
    BN_set_word(bne, RSA_F4);
    RSA *r = RSA_new();
    RSA_generate_key_ex(r, 2048, bne, nullptr);
    EVP_PKEY_assign_RSA(pkey, r);

    // generating a certificate
    ASN1_INTEGER_set(X509_get_serialNumber(cert), 1);
    X509_gmtime_adj(X509_get_notBefore(cert), 0);
    X509_gmtime_adj(X509_get_notAfter(cert), 31536000L);
    X509_set_pubkey(cert, pkey);

    // setting certificate information
    X509_NAME *name = X509_get_subject_name(cert);
    X509_NAME_add_entry_by_txt(name, "C", MBSTRING_ASC, (const unsigned char *)"CN", -1, -1, 0);
    X509_NAME_add_entry_by_txt(name, "O", MBSTRING_ASC, (const unsigned char *)"My Company", -1, -1, 0);
    X509_NAME_add_entry_by_txt(name, "CN", MBSTRING_ASC, (const unsigned char *)"My Root CA", -1, -1, 0);
    X509_set_issuer_name(cert, caCert != nullptr ? X509_get_subject_name(caCert) : name);

    // add subject alt name in extension
    X509_EXTENSION *ext = X509V3_EXT_conf_nid(nullptr, nullptr, NID_subject_alt_name, "DNS:ServiceDNS");
    X509_add_ext(cert, ext, -1);

    // signature certificate
    X509_sign(cert, caPkey != nullptr ? caPkey : pkey, EVP_sha256());

    // releasing memory
    X509_EXTENSION_free(ext);
    name = nullptr;
    r = nullptr;
    BN_free(bne);
}

static void GeneratePrivateKey(const std::string &password)
{
    EVP_PKEY *pkey = NULL;
    EVP_PKEY_CTX *ctx = EVP_PKEY_CTX_new_id(EVP_PKEY_RSA, NULL);
    if (!ctx) {
        std::cerr << "Failed to create EVP_PKEY_CTX" << std::endl;
        return;
    }
    if (EVP_PKEY_keygen_init(ctx) <= 0) {
        std::cerr << "Failed to initialize key generation" << std::endl;
        EVP_PKEY_CTX_free(ctx);
        return;
    }
    if (EVP_PKEY_CTX_set_rsa_keygen_bits(ctx, 2048) <= 0) {
        std::cerr << "Failed to set key length" << std::endl;
        EVP_PKEY_CTX_free(ctx);
        return;
    }
    if (EVP_PKEY_keygen(ctx, &pkey) <= 0) {
        std::cerr << "Failed to generate key" << std::endl;
        EVP_PKEY_CTX_free(ctx);
        return;
    }
    EVP_PKEY_CTX_free(ctx);
    BIO *bio = BIO_new(BIO_s_mem());
    if (!bio) {
        std::cerr << "Failed to create BIO" << std::endl;
        EVP_PKEY_free(pkey);
        return;
    };
    if (!PEM_write_bio_PrivateKey(bio, pkey, EVP_aes_256_cbc(), (unsigned char *)password.c_str(), (int)password.size(),
                                  NULL, NULL)) {
        std::cerr << "Failed to write private key" << std::endl;
        BIO_free_all(bio);
        EVP_PKEY_free(pkey);
        return;
    }
    BUF_MEM *buf;
    BIO_get_mem_ptr(bio, &buf);
    std::ofstream outfile("/tmp/cert.key", std::ios::out | std::ios::binary);
    if (!outfile) {
        std::cerr << "Failed to open file" << std::endl;
        BIO_free_all(bio);
        EVP_PKEY_free(pkey);
        return;
    }
    outfile.write(buf->data, buf->length);
    outfile.close();
    BIO_free_all(bio);
    EVP_PKEY_free(pkey);
}

TEST_F(CertsUtilsTest, GetPemCertsFromFilesTest)
{
    // gen ca cert
    EVP_PKEY *caPkey = EVP_PKEY_new();
    X509 *caCert = X509_new();
    GenPemCert(caPkey, caCert);
    // gen pem cert
    EVP_PKEY *pkey = EVP_PKEY_new();
    X509 *x509 = X509_new();
    GenPemCert(pkey, x509, caPkey, caCert);

    std::string caFile = "/tmp/ca.crt";
    FILE *fCa = fopen(caFile.c_str(), "wb");
    PEM_write_X509(fCa, x509);
    fclose(fCa);
    std::string certFile = "/tmp/cert.crt";
    FILE *fCert = fopen(certFile.c_str(), "wb");
    PEM_write_X509(fCert, x509);
    fclose(fCert);

    GeneratePrivateKey("123456");

    ASSERT_TRUE(ExistPath("/tmp/cert.key") && ExistPath("/tmp/cert.crt") && ExistPath("/tmp/ca.crt"));

    EVP_PKEY *privatekey = GetPrivateKeyFromFile("/tmp/cert.key", "123456");
    X509 *publicKey = GetCertFromFile("/tmp/cert.crt");
    STACK_OF(X509) *rootCaCerts = GetCAFromFile("/tmp/ca.crt");

    auto prikey = GetPrivateKey(privatekey);
    auto cert = GetCert(publicKey);
    auto ca = GetCa(rootCaCerts);
    EXPECT_TRUE(!prikey.Empty());
    EXPECT_TRUE(!cert.empty());
    EXPECT_TRUE(!ca.empty());
    // releasing memory
    X509_free(x509);
    EVP_PKEY_free(pkey);
    X509_free(caCert);
    EVP_PKEY_free(caPkey);
    prikey.Clear();
    ClearPemCerts(privatekey, publicKey, rootCaCerts);
    Rm("/tmp/cert.crt");
    Rm("/tmp/ca.crt");
    Rm("/tmp/cert.key");
}
}  // namespace test
}  // namespace YR
