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

package org.yuanrong.jni;

import lombok.Data;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Init environment information
 *
 * @since 2024/10/11
 */
@Data
public class YRAutoInitInfo {
    private String functionSystemServerIpAddr = "";
    private int functionSystemServerPort;
    private String dataSystemIpAddr = "";
    private int dataSystemPort;

    private String serverAddr = "";
    private String dsAddr = "";
    private boolean inCluster;

    /**
     * The constructor of YRAutoInitInfo.
     *
     * @param functionSystemServerAddr the function system server address
     * @param dataSystemAddr the data system address
     * @param inCluster inCluster
     */
    public YRAutoInitInfo(String functionSystemServerAddr, String dataSystemAddr, boolean inCluster) {
        this.inCluster = inCluster;
        this.setServerAddr(functionSystemServerAddr);
        this.setDsAddr(dataSystemAddr);
    }

    /**
     * The constructor of YRAutoInitInfo.
     */
    public YRAutoInitInfo() {}

    public final void setServerAddr(String serverAddr) {
        this.serverAddr = serverAddr;
        String addrRegex = "^(?:http://|https://|grpc://)?(.*):(\\d+)$";
        Pattern pattern = Pattern.compile(addrRegex);
        Matcher serverAddrMatcher = pattern.matcher(serverAddr);
        if (serverAddrMatcher.matches()) {
            this.functionSystemServerIpAddr = serverAddrMatcher.group(1);
            this.functionSystemServerPort = Integer.parseInt(serverAddrMatcher.group(2));
        }
    }

    public void setDsAddr(String dsAddr) {
        this.dsAddr = dsAddr;
        String addrRegex = "^(?:http://|https://|grpc://)?(.*):(\\d+)$";
        Pattern pattern = Pattern.compile(addrRegex);
        Matcher dsAddrMatcher = pattern.matcher(dsAddr);
        if (dsAddrMatcher.matches()) {
            this.dataSystemIpAddr = dsAddrMatcher.group(1);
            this.dataSystemPort = Integer.parseInt(dsAddrMatcher.group(2));
        }
    }
}
