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

package org.yuanrong.api;

import lombok.Data;

import java.util.HashSet;

/**
 * the InvokeArg
 *
 * @since 2023/11/6
 */
@Data
public class InvokeArg {
    private byte[] data;

    private boolean isObjectRef = false;

    private String objId;

    private HashSet<String> nestedObjects;

    /**
     * Init InvokeArg
     */
    public InvokeArg() {}

    /**
     *  Init InvokeArg
     *
     * @param inputBytes the inputBytes
     */
    public InvokeArg(byte[] inputBytes) {
        this.data = inputBytes;
    }

    /**
     * set the isObjectRef
     *
     * @param isObjectRef the isObjectRef
     */
    public void setObjectRef(boolean isObjectRef) {
        this.isObjectRef = isObjectRef;
    }

    /**
     * set the nestedObjects
     *
     * @param nestedObjects the nestedObjects hashSet
     */
    public void setNestedObjects(HashSet<String> nestedObjects) {
        this.nestedObjects = nestedObjects;
    }

    /**
     * get the data
     *
     * @return byte[]
     */
    public byte[] getData() {
        return this.data;
    }

    /**
     * get the isObjectRef
     *
     * @return the isObjectRef
     */
    public boolean isObjectRef() {
        return this.isObjectRef;
    }

    /**
     * get the objId
     *
     * @return the objId
     */
    public String getObjectId() {
        return this.objId;
    }

    /**
     * set the objId
     *
     * @param objId the objId
     */
    public void setObjId(String objId) {
        this.objId = objId;
    }

    /**
     * get the nestedObjects
     *
     * @return the nestedObjects
     */
    public HashSet<String> getNestedObjects() {
        return this.nestedObjects;
    }
}