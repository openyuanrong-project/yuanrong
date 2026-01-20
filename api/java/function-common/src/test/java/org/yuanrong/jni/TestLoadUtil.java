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

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import org.junit.Test;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class TestLoadUtil {
    @Test
    public void testLoadLibraryWithJvmParam() throws IOException {
        LoadUtil loadUtil = new LoadUtil();
        String exceptionMsg = "";
        try {
            LoadUtil.loadLibrary();
        } catch (ExceptionInInitializerError e) {
            exceptionMsg = e.getCause().toString();
        }
        System.out.println(exceptionMsg);
        // LoadLibraryWithJvmParam failed, then finding the so.properties also failed.
        assertTrue(exceptionMsg.contains("FileNotFoundException"));
    }

    @Test
    public void testCheckSHA256() throws IOException {
        String filePath = "/tmp/sha256test.txt";

        createFile(filePath);
        Process process = Runtime.getRuntime().exec("sha256sum " + filePath);
        BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()));
        String line = reader.readLine();
        String[] parts = line.split(" ");
        String sha256 = parts[0];
        reader.close();

        File file = Paths.get(filePath).toFile();
        assertTrue(LoadUtil.checkSHA256(file, sha256));
        file.deleteOnExit();
    }

    @Test
    public void testCheckSHA256faied() throws IOException {
        String filePath = "/tmp/sha256testfailed.txt";
        createFile(filePath);

        String sha256 = "sha256fromproperties";

        File file = Paths.get(filePath).toFile();
        assertFalse(LoadUtil.checkSHA256(file, sha256));
        file.deleteOnExit();
    }

    private void createFile(String filePath) throws IOException {
        Path path = Paths.get(filePath);
        Files.createDirectories(path.getParent());

        try (FileOutputStream fos = new FileOutputStream(path.toFile())) {
            fos.write("".getBytes(StandardCharsets.UTF_8));
        }
    }
}
