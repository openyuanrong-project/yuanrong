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

package org.yuanrong.utils;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.io.IOException;
import java.lang.reflect.Field;
import java.util.Map;
import java.nio.ByteBuffer;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

/**
 * Runtime utils.
 *
 * @since 2024 /7/21
 */
public class RuntimeUtils {
    private static final Logger LOGGER = LogManager.getLogger(RuntimeUtils.class);

    /**
     * Helper function to convert arglist to string List.
     *
     * @param argList Arg list
     * @return a list of string
     */
    public static List<String> convertArgListToStringList(List<ByteBuffer> argList) {
        List<String> args = new ArrayList<>();
        for (ByteBuffer arg : argList) {
            Charset charset = Charset.forName("UTF-8");
            args.add(charset.decode(arg).toString());
        }
        return args;
    }

    /**
     * If YR_SEED_FILE is set, read the specified file to block.
     */
    public static void readSeedFile() {
        String seedFile = System.getenv("YR_SEED_FILE");
        LOGGER.info("Begin to read seed file: {}", seedFile);
        if (seedFile == null || seedFile.trim().isEmpty()) {
            return;
        }
        try {
            Files.readAllBytes(Paths.get(seedFile));
        } catch (IOException e) {
            LOGGER.error("Failed to read seed file: {}", seedFile, e);
        }
    }


    /**
     * Load KEY=value lines from a file into JVM system properties (same keys as setJavaProcessEnv).
     *
     * @param envFilePath path to env file; empty or missing file is ignored with log
     */
    public static void loadEnvFromFile(String envFilePath) {
        if (envFilePath == null || envFilePath.trim().isEmpty()) {
            LOGGER.info("Environment file is empty");
            return;
        }
        LOGGER.info("Loading env from file: {}", envFilePath);
        Path path = Paths.get(envFilePath);
        if (!Files.exists(path)) {
            LOGGER.warn("Environment variable file not found: {}", envFilePath);
            return;
        }

        int loaded = 0;
        try {
            List<String> lines = Files.readAllLines(path, StandardCharsets.UTF_8);
            for (int lineNum = 1; lineNum <= lines.size(); lineNum++) {
                String line = lines.get(lineNum - 1).trim();
                if (shouldSkip(line)) {
                    continue;
                }
                String[] kv = parseLine(line, lineNum, envFilePath);
                if (kv.length == 0) {
                    continue;
                }
                // Write into JVM system properties (globally visible)
                setJavaProcessEnv(kv[0], kv[1]);
                loaded++;
            }
        } catch (IOException e) {
            LOGGER.error("Failed to load environment variables from {}", envFilePath, e);
            return;
        }
        if (loaded > 0) {
            LOGGER.debug("Loaded {} environment variables from {}", loaded, envFilePath);
        }
    }

    private static boolean shouldSkip(String line) {
        return line.isEmpty() || line.startsWith("#");
    }

    private static String[] parseLine(String line, int lineNum, String filePath) {
        // Separate by the first "=" sign.
        int idx = line.indexOf('=');
        if (idx == -1) {
            LOGGER.warn("{}:{}  invalid line, missing '=': {}", filePath, lineNum, line);
            return new String[0];
        }
        String key = line.substring(0, idx).trim();
        // Keep the spaces, no quotes.
        String value = line.substring(idx + 1).trim();
        if (key.isEmpty()) {
            LOGGER.warn("{}:{}  empty key", filePath, lineNum);
            return new String[0];
        }
        return new String[]{key, value};
    }

    /**
     * By default, get from env, is missing, get from jvm property, if both failed,
     * log error and return empty string
     *
     * @param key 要获取的环境变量或系统属性键，不能为 null
     * @return 对应的值，如果都无法获取则返回空字符串
     * @throws IllegalArgumentException if key is null
     */
    public static String getEnvOrProperty(String key) {
        if (key == null) {
            throw new IllegalArgumentException("Key cannot be null");
        }
        // attempt to obtain it from the environment variables.
        String envValue = System.getenv(key);
        if (envValue != null && !envValue.isEmpty()) {
            return envValue;
        }
        // If the environment variable does not exist or is empty, attempt to obtain it from the system properties.
        String propertyValue = System.getProperty(key);
        if (propertyValue != null && !propertyValue.isEmpty()) {
            return propertyValue;
        }
        return "";
    }

    /**
     * Set environment variable to process environment map, accessible via System.getenv().
     *
     * @param key the environment variable key
     * @param value the environment variable value
     */
    public static void setJavaProcessEnv(String key, String value) {
        // keep process previous env map configs
        try {
            Class<?> processEnvironmentClass = Class.forName("java.lang.ProcessEnvironment");
            updateJavaEnvMap(processEnvironmentClass, "theCaseInsensitiveEnvironment", key, value);
            updateJavaEnvMap(processEnvironmentClass, "theUnmodifiableEnvironment", key, value);
        } catch (ClassNotFoundException e) {
            LOGGER.error("get field: theEnvironment has an error: ", e);
        }
    }

    /**
     * Update java env map.
     *
     * @param cls the cls
     * @param fieldName the field name
     * @param key the key
     * @param value the value
     */
    private static void updateJavaEnvMap(Class<?> cls, String fieldName, String key, String value) {
        try {
            // get field and access
            Field oldField = cls.getDeclaredField(fieldName);
            oldField.setAccessible(true);
            // get Field map
            Object unmodifiableMap = oldField.get(null);
            injectIntoUnmodifiableMap(key, value, unmodifiableMap);
        } catch (ReflectiveOperationException e) {
            LOGGER.error("get field: {} has an error: {}", fieldName, e);
        }
    }

    private static void injectIntoUnmodifiableMap(String key, String value, Object map)
            throws ReflectiveOperationException {
        Class<?> unmodifiableMapClass = Class.forName("java.util.Collections$UnmodifiableMap");
        Field field = unmodifiableMapClass.getDeclaredField("m");
        field.setAccessible(true);
        Object obj = field.get(map);
        ((Map<String, String>) obj).put(key, value);
    }
}
