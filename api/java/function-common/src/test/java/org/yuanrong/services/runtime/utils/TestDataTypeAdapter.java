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

package org.yuanrong.services.runtime.utils;

import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;

import org.junit.Assert;
import org.junit.Test;
import org.mockito.Mockito;

import java.io.IOException;
import java.io.StringReader;
import java.io.StringWriter;
import java.util.List;
import java.util.Map;

public class TestDataTypeAdapter {
    private DataTypeAdapter adapter = new DataTypeAdapter();

    @Test
    public void testReadArray() throws IOException {
        String json = "[1, 2, 3]";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertTrue(result instanceof List);
        Assert.assertEquals(3, ((List) result).size());
    }

    @Test
    public void testReadObject() throws IOException {
        String json = "{\"key\": \"value\"}";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertTrue(result instanceof Map);
        Assert.assertEquals("value", ((Map) result).get("key"));
    }

    @Test
    public void testReadString() throws IOException {
        String json = "\"value\"";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertTrue(result instanceof String);
        Assert.assertEquals("value", result);
    }

    @Test
    public void testReadNumber() throws IOException {
        String json = "123";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertTrue(result instanceof Integer);
        Assert.assertEquals(123, result);
    }

    @Test
    public void testReadBoolean() throws IOException {
        String json = "true";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertTrue(result instanceof Boolean);
        Assert.assertEquals(true, result);
    }

    @Test
    public void testReadNull() throws IOException {
        String json = "null";
        JsonReader reader = new JsonReader(new StringReader(json));
        Object result = adapter.read(reader);
        Assert.assertNull(result);
    }

    @Test
    public void testWrite() throws IOException {
        JsonWriter out = new JsonWriter(new StringWriter());
        Object value = new Object();
        DataTypeAdapter adapter = Mockito.mock(DataTypeAdapter.class);
        adapter.write(out, value);
        Mockito.verify(adapter).write(out, value);
    }
}
