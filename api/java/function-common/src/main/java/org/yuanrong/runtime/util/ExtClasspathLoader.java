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

package org.yuanrong.runtime.util;

import org.apache.commons.io.FileUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * get jar path from env and load all of them
 *
 * @since 2022/7/27
 */
public final class ExtClasspathLoader {
    private static Method addURL;

    private static final String HOME_PATH = "/dcache";

    private static final String JAR_STRING = ".jar";

    private static final String LINE_BREAK = System.lineSeparator();

    private static final Logger LOG = LoggerFactory.getLogger(ExtClasspathLoader.class);

    private static final URLClassLoader CLASSLOADER = new FuncClassLoader(new URL[] {},
        ClassLoader.getSystemClassLoader().getParent());

    static {
        try {
            addURL = URLClassLoader.class.getDeclaredMethod("addURL", new Class[] {URL.class});
            addURL.setAccessible(true);
        } catch (SecurityException | NoSuchMethodException e) {
            LOG.error("fail to init addMethod {}", e.getMessage());
        }
    }

    /**
     * load jar classpath。
     *
     * @param codePaths the codePaths
     * @throws IOException throw load user function Failed
     * @throws InvocationTargetException throw load user function Failed
     * @throws IllegalAccessException    throw load user function Failed
     */
    public static void loadClasspath(List<String> codePaths)
        throws IOException, InvocationTargetException, IllegalAccessException {
        codePaths.addAll(getJarFiles());
        LOG.debug("function lib path: {}", codePaths);
        for (String filepath : codePaths) {
            if (filepath == null || filepath.isEmpty()) {
                continue;
            }
            File file = FileUtils.getFile(clearFilePathString(filepath));
            if (!file.exists()) {
                LOG.warn("the file: {} does not exist.", filepath);
                continue;
            }
            loopFiles(file);
        }
    }

    private static String clearFilePathString(String path) {
        return path.replace(LINE_BREAK, "");
    }

    /**
     * Traverse dirs recursively to find all JAR packages under this dir.
     *
     * @param file dir that need to search
     * @throws InvocationTargetException throw loop dir for user function Failed
     * @throws IllegalAccessException    throw loop dir for user function Failed
     */
    private static void loopFiles(File file) throws InvocationTargetException, IllegalAccessException {
        if (file == null) {
            return;
        }
        if (file.isDirectory()) {
            File[] fileList = file.listFiles();
            LOG.debug("file list: {}", Arrays.toString(fileList));
            if (fileList == null) {
                LOG.error("null pointer exception when listing files {}", file);
                throw new IllegalArgumentException("null pointer exception when listing files");
            }
            for (File tmp : fileList) {
                loopFiles(tmp);
            }
        } else {
            try {
                if (isEndWithJar(file.getPath())) {
                    LOG.debug("load jar file {}", file.getName());
                    addURL(file);
                }
            } catch (IOException e) {
                LOG.error("fail to get canonical path of jar file:", e);
                throw new IllegalArgumentException("fail to get canonical path of jar file");
            }
        }
    }

    /**
     * load files to classPath via filepath.
     *
     * @param file file path
     * @throws IllegalAccessException    addURL Failed to load class due
     *                                   IllegalAccessException
     * @throws IllegalArgumentException  addURL Failed to load class due
     *                                   IllegalArgumentException
     * @throws MalformedURLException     addURL Failed to load class due
     *                                   MalformedURLException
     * @throws InvocationTargetException addURL Failed to load class due
     *                                   InvocationTargetException
     */
    private static void addURL(File file)
        throws MalformedURLException, InvocationTargetException, IllegalAccessException {
        try {
            if (file == null) {
                LOG.error("input file is illegal");
                throw new IllegalArgumentException("input file is illegal");
            }
            addURL.invoke(CLASSLOADER, file.toURI().toURL());
        } catch (IllegalAccessException | IllegalArgumentException | InvocationTargetException
                 | MalformedURLException e) {
            LOG.error("addURL failed for file {} due to exception {}", file, e);
            throw e;
        }
    }

    /**
     * load file path from environment variable "YR_FUNCTION_LIB_PATH"
     *
     * @return jar file paths
     */
    private static List<String> getJarFiles() {
        String pathFromEnv = System.getenv("YR_FUNCTION_LIB_PATH");
        if (pathFromEnv == null || pathFromEnv.isEmpty()) {
            LOG.warn("failed to get YR_FUNCTION_LIB_PATH variable, it may be ok");
        }
        List<String> libPath = new ArrayList<String>(2);
        libPath.add(0, pathFromEnv);

        String delegateDownload = System.getenv("ENV_DELEGATE_DOWNLOAD");
        if (delegateDownload == null || delegateDownload.isEmpty()) {
            LOG.warn("failed to get ENV_DELEGATE_DOWNLOAD variable, it may be ok");
        } else {
            libPath.add(1, delegateDownload);
        }
        return libPath;
    }

    /**
     * Check if path starts with /dcache.
     *
     * @param path path to check
     * @return true if path starts with /dcache otherwise false
     */
    public static boolean isStartWithHomePath(String path) {
        return path.startsWith(HOME_PATH);
    }

    /**
     * Check if path ends with .jar
     *
     * @param path file path to check
     * @return true if path ends with .jar otherwise false
     */
    public static boolean isEndWithJar(String path) {
        return path.endsWith(JAR_STRING);
    }

    /**
     * getter of function CLASSLOADER
     *
     * @return CLASSLOADER
     */
    public static ClassLoader getFunctionClassLoader() {
        return CLASSLOADER;
    }
}
