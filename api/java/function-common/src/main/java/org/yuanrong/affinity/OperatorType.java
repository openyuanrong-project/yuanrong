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

package org.yuanrong.affinity;

/**
 * Description:
 *
 * @since 2023-11-26
 */
public enum OperatorType {
    LABEL_IN(1),
    LABEL_NOT_IN(2),
    LABEL_EXISTS(3),
    LABEL_NOT_EXISTS(4);

    private int type;

    OperatorType(int type) {
        this.type = type;
    }

    int getType() {
        return type;
    }
}
