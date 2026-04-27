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

package org.yuanrong.executor;

import org.yuanrong.services.runtime.Context;

import com.google.gson.Gson;
import com.google.gson.JsonObject;

public class UserTestHandler {
    public String handler(TestRequestEvent event, Context context) {
        Gson gson = new Gson();
        System.out.println(gson.toJson(event));
        if (event.getLevel() == 0) {
            int c = 100 / event.getLevel();
        }
        return "true";
    }

    public String timeoutHandler(TestRequestEvent event, Context context) {
        try {
            Thread.sleep(3 * 1000 + 500);
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }
        return "true";
    }

    public String contextHandler(String key, Context context) {
        return context.getUserData(key);
    }

    public String failedHandler(String event, Context context) {
        throw new RuntimeException("execute handler failed");
    }

    public String instanceLabelHandler(TestRequestEvent event, Context context) {
        return context.getInstanceLabel();
    }

    public void initializer(Context context) {
        System.out.println("initialize success");
    }

    public void illegalArgumentInitializer(Context context) {
        throw new IllegalArgumentException("initialize failed");
    }

    public void failedInitializer(Context context) {
        throw new RuntimeException("initialize failed");
    }

    public void preStop(Context context) {
        System.out.println("preStop success");
    }

    public void failedPreStop(Context context) {
        throw new RuntimeException("preStop failed");
    }

    public void timeoutPreStop(Context context) throws InterruptedException {
        Thread.sleep(2000);
    }

    public void timeoutInitializer(Context context) {
        try {
            Thread.sleep(3 * 1000 + 500);
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }
    }

    public String largeResponse(TestRequestEvent event, Context context) {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < 1024 * 1024; i++) {
            sb.append("aaaaaaaaaa");
        }
        return sb.toString();
    }

    public String jsonHandler(JsonObject jsonObject, Context context) {
        Gson gson = new Gson();
        return gson.toJson(jsonObject);
    }
}
