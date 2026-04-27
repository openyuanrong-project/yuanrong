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

import org.yuanrong.exception.YRException;

/**
 * SDK interface for reading the current invocation's session.
 *
 * <p>Session creation, locking, and persistence are all managed by the
 * runtime automatically. User code only needs to call {@link #loadSession()}
 * to obtain the current {@link SessionObj}.</p>
 *
 * @since 2026/03/25
 */
public interface SessionService {
    /**
     * Load the session associated with the current invocation.
     *
     * <p>Returns {@code null} when no {@code sessionId} is present in the
     * request or {@code use_agent_session=false}.</p>
     *
     * @return the current {@link SessionObj}, or {@code null}
     * @throws YRException if the native call fails
     */
    SessionObj loadSession() throws YRException;
}
