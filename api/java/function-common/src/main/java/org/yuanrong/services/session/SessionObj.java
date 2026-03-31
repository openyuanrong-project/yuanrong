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

import java.util.List;

/**
 * Represents an agent session object.
 *
 * <p>The session object is managed by the runtime. Users should only read and
 * modify the session via the provided accessors. Do NOT modify the list returned
 * by {@link #getHistories()} directly; always use {@link #setHistories(List)} to
 * write back changes so the runtime can be notified.</p>
 *
 * @since 2026/03/25
 */
public interface SessionObj {
    /**
     * Get the session ID.
     *
     * @return session ID
     */
    String getID();

    /**
     * Get a read-only snapshot of the conversation history.
     *
     * <p>The returned list is a defensive copy / unmodifiable view.
     * Mutating it has no effect on the runtime state.</p>
     *
     * @return unmodifiable list of history entries
     */
    List<String> getHistories();

    /**
     * Update the conversation history.
     *
     * <p>This method immediately synchronizes the new value to libruntime
     * via JNI so that the runtime holds the latest state before persisting.</p>
     *
     * @param histories new list of history entries (must not be null)
     * @throws YRException if the JNI call fails
     */
    void setHistories(List<String> histories) throws YRException;

    /**
     * Check whether the session has been interrupted.
     *
     * <p>This method queries libruntime to determine if the current session
     * has been interrupted (e.g., by a cancellation request).</p>
     *
     * @return true if the session has been interrupted, false otherwise
     * @throws YRException if the JNI call fails
     */
    boolean isInterrupted() throws YRException;
}
