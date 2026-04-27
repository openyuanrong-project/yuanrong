/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

package org.yuanrong.testutils;

import java.io.File;
import java.io.PrintWriter;
import java.io.StringWriter;

import org.yuanrong.Config;
import org.yuanrong.Config.Builder;
import org.yuanrong.api.ClientInfo;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;

public class TestUtils {
    public static void initYR(boolean withUrn) throws YRException {
        Config conf = createConfig(withUrn);
        printConfigInfo(conf);
        YR.init(conf);
    }

    public static void initYR() throws YRException {
        initYR(true);
    }

    public static String initYRWithTenantCtx(boolean enableSetContext, boolean isThreadLocal)
            throws YRException {
        Config conf = createConfig(true);
        conf.setThreadLocal(isThreadLocal);
        conf.setEnableSetContext(enableSetContext);
        printConfigInfo(conf);
        ClientInfo info = YR.init(conf);
        return info.getContext();
    }

    public static void initYRInvalid() throws YRException {
        Config conf = createInvalidConfig();
        printConfigInfo(conf);
        YR.init(conf);
    }

    public static String getStackTraceStr(Exception e) {
        StringWriter stringWriter = new StringWriter();
        PrintWriter printWriter = new PrintWriter(stringWriter);
        e.printStackTrace(printWriter);
        String causedBy = stringWriter.toString();
        printWriter.close();
        return causedBy;
    }

    public static Config createConfig(boolean withUrn) {
        String functionURN = System.getenv("YR_JAVA_FUNC_ID");
        String serverIpPort = System.getenv("YR_SERVER_ADDRESS");
        String dsIpPort = System.getenv("YR_DS_ADDRESS");
        String deployPath = System.getenv("DEPLOY_PATH");
        String logLevel = System.getenv("YR_LOG_LEVEL");
        String logDir = System.getenv("GLOG_log_dir");;
        String[] serverParts = serverIpPort.split(":");
        String[] dsParts = dsIpPort.split(":");

        Builder tempConfig = Config.builder()
                .serverAddress(serverParts[0])
                .serverAddressPort(Integer.parseInt(serverParts[1]))
                .dataSystemAddress(dsParts[0])
                .dataSystemAddressPort(Integer.parseInt(dsParts[1]))
                .logDir(logDir)
                .logLevel(logLevel);
        if (withUrn) {
            tempConfig.functionURN(functionURN);
        }
        return tempConfig.build();
    }

    public static Config createInvalidConfig() {
        String functionURN = System.getenv("YR_JAVA_FUNC_ID");
        Config conf = new Config(functionURN, "127.0.0.1", "", "", true);
        return conf;
    }

    private static void printConfigInfo(Config config) {
        StringBuilder strbuilder = new StringBuilder();
        String lineSept = System.lineSeparator();
        String str = strbuilder
                .append("[INFO][initYR] Init YR with:")
                .append(lineSept + "[INFO][initYR]    functionID: " + config.getFunctionURN())
                .append(lineSept + "[INFO][initYR]    serverAddress: " + config.getServerAddress() + ":"
                        + Integer.toString(config.getServerAddressPort()))
                .append(lineSept + "[INFO][initYR]    dsAddress: " + config.getDataSystemAddress() + ":"
                        + Integer.toString(config.getDataSystemAddressPort()))
                .append(lineSept + "[INFO][initYR]    logDir: " + config.getLogDir())
                .toString();

        System.out.println(str);
    }

    public static Integer returnInt(int num) {
        return num;
    }

    public static String returnCustomEnvs(String key) {
        String env = System.getenv(key);
        if (env != null) {
            System.out.println("custom env: " + env);
            return env;
        } else {
            return "";
        }
    }
}
