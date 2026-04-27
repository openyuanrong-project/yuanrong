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

package org.yuanrong;

import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;

import lombok.Data;

/**
 * Group
 * <p>
 * Public API.
 * Please note that it should not be modified without careful consideration.
 * Any modifications require strict review and testing.
 * </p>
 *
 * @since 2024/04/15
 */
@Data
public class Group {
    private String groupName;
    private GroupOptions groupOpts;

    /**
     * Non-parameter constructor.
     */
    public Group() {}

    /**
     * Group constructor.
     *
     * @param name the name.
     * @param opts the opts.
     */
    public Group(String name, GroupOptions opts) {
        this.groupOpts = opts;
        this.groupName = name;
    }

    /**
     * Deprecated. Group scheduling instance invoke interface, group instance creation.
     *
     * @throws YRException Group instances follow the Fate-sharing principle. If a certain instance fails to be
     *                            created, the interface throws an exception, and all instances fail to be created.
     */
    @Deprecated
    public void Invoke() throws YRException {
        invoke();
    }

    /**
     * Group scheduling instance invoke interface, group instance creation.
     *
     * @note 1.A single group can create up to 256 instances.\n
     *       2.Concurrent creation supports up to 12 groups, and each group can create up to 256 instances.\n
     *       3.Calling invoke() after NamedInstance::Export() causes the current thread to hang.\n
     *       4.If you do not call the invoke () interface and directly send a function request to a stateful function
     *       instance and get the result, the current thread will be blocked.\n
     *       5.Repeatedly calling the invoke() interface will cause an exception to be thrown.\n
     *       6.Instances within a group do not support specifying a detached lifecycle.\n
     *
     * @throws YRException Group instances follow the Fate-sharing principle. If a certain instance fails to be
     *                            created, the interface throws an exception, and all instances fail to be created.
     *
     * @snippet{trimleft} GroupExample.java Group invoke example
     */
    public void invoke() throws YRException {
        YR.getRuntime().groupCreate(groupName, groupOpts);
    }

    /**
     * Deprecated. Group scheduling instance Terminate interface, delete group instance creation.
     *
     * @throws YRException the YR exception.
     */
    @Deprecated
    public void Terminate() throws YRException {
        terminate();
    }

    /**
     * Group scheduling instance terminate interface, delete group instance creation.
     *
     * @throws YRException the YR exception.
     *
     * @snippet{trimleft} GroupExample.java Group terminate example
     */
    public void terminate() throws YRException {
        YR.getRuntime().groupTerminate(groupName);
    }

    /**
     * Group scheduling instance Wait interface, wait for the group instance creation to complete.
     *
     * @throws YRException Group instances follow the Fate-sharing principle. If a certain instance fails to be
     *                            created, the interface throws an exception, and all instances fail to be created.
     *                            The interface throws an exception after the wait timeout.
     */
    public void Wait() throws YRException {
        YR.getRuntime().groupWait(groupName);
    }
}

