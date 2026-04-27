/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.
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

package org.yuanrong.services.session;

import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.services.session.ManagedSessionObj;

/**
 * Default implementation of {@link SessionService}.
 *
 * <p>Calls the native {@code LoadCurrentSession(sessionId)} bridge to read
 * the session that libruntime already loaded (via {@code AcquireInvokeSession})
 * before the user function was invoked. DataSystem is NOT accessed directly.</p>
 *
 * @since 2026/03/25
 */
public class SessionServiceImpl implements SessionService {
    private final String sessionId;

    public SessionServiceImpl(String sessionId) {
        this.sessionId = sessionId == null ? "" : sessionId;
    }

    /**
     * Load the session for this invocation.
     *
     * <p>Returns {@code null} when {@code sessionId} is empty, meaning no session
     * was associated with this request or {@code use_agent_session} is disabled.</p>
     *
     * @return {@link ManagedSessionObj} wrapping the current session, or {@code null}
     * @throws YRException if the native call fails
     */
    @Override
    public SessionObj loadSession() throws YRException {
        if (sessionId.isEmpty()) {
            return null;
        }
        try {
            String json = LibRuntime.loadCurrentSession(sessionId);
            return ManagedSessionObj.fromJson(json);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
    }
}
