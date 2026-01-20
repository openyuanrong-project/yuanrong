/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.jobexecutor;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import lombok.Getter;

import java.util.HashMap;
import java.util.Map;

/**
 * The OBS(Object Storage Service) configuration for remote JobExecutor runtime
 * to download driver code.
 *
 * @since 2023 /11/11
 */
@Getter
public class OBSoptions {
    private static final String DELEGATE_DOWNLOAD = "DELEGATE_DOWNLOAD";

    private static final String BUCKET_ID = "bucketId";

    private static final String OBJECT_ID = "objectId";

    private static final String HOSTNAME = "hostName";

    private static final String SECURITY_TOKEN = "securityToken";

    private static final String TEMPORARY_ACCESS_KEY = "temporaryAccessKey";

    private static final String TEMPORARY_SECRET_KEY = "temporarySecretKey";

    private String endPoint = "";

    private String bucketID = "";

    private String objectID = "";

    private String securityToken = "";

    private String ak = "";

    private String sk = "";

    /**
     * The access address of OBS.
     *
     * @param endPoint the access address String.
     */
    public void setEndPoint(String endPoint) {
        this.endPoint = endPoint;
    }

    /**
     * The container ID for storing objects in OBS.
     *
     * @param bucketID the container ID String.
     */
    public void setBucketID(String bucketID) {
        this.bucketID = bucketID;
    }

    /**
     * The ID of OBS object to be downloaded.
     *
     * @param objectID the object ID String.
     */
    public void setObjectID(String objectID) {
        this.objectID = objectID;
    }

    /**
     * The access token issued by the system to an IAM user. It carries information
     * such as user identities and permissions.
     *
     * @param securityToken the security token String.
     */
    public void setSecurityToken(String securityToken) {
        this.securityToken = securityToken;
    }

    /**
     * The temporary access key(AK) to OBS. It is a unique identifier associated
     * with the secret access key(SK). AK and SK are used together to encrypt and
     * sign requests.
     *
     * @param ak the temporary access key String.
     */
    public void setAk(String ak) {
        this.ak = ak;
    }

    /**
     * The temporary secret access key(SK) used together with the temporary access
     * key(AK) to encrypt and sign requests to OBS.
     *
     * @param sk the temporary secret access key String.
     */
    public void setSk(String sk) {
        this.sk = sk;
    }

    /**
     * Converts OBS options to a map contains an single item { "DELEGATE_DOWNLOAD":
     * jsonString }, which can be put into createOptions.
     *
     * @return Map<String, String> map contains String key and jsonString value.
     * @throws YRException Failed to convert mapper to json string.
     */
    public Map<String, String> toMap() throws YRException {
        ObjectMapper mapper = new ObjectMapper();
        ObjectNode rootNode = mapper.createObjectNode();
        rootNode.put(HOSTNAME, endPoint);
        rootNode.put(BUCKET_ID, bucketID);
        rootNode.put(OBJECT_ID, objectID);
        rootNode.put(SECURITY_TOKEN, securityToken);
        rootNode.put(TEMPORARY_ACCESS_KEY, ak);
        rootNode.put(TEMPORARY_SECRET_KEY, sk);

        String jsonStr;
        try {
            jsonStr = mapper.writeValueAsString(rootNode);
        } catch (JsonProcessingException e) {
            throw new YRException(ErrorCode.ERR_JOB_USER_CODE_EXCEPTION, ModuleCode.RUNTIME,
                    "Failed to convert OBS options to a json string.");
        }
        Map<String, String> map = new HashMap<>();
        map.put(DELEGATE_DOWNLOAD, jsonStr);
        return map;
    }
}