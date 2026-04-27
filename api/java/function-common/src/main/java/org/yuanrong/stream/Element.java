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

package org.yuanrong.stream;

import lombok.Data;
import lombok.experimental.Accessors;

import java.nio.ByteBuffer;
import java.util.Objects;

/**
 * Element class that contains element id and data cache.
 *
 * @since 2024/09/04
 */
@Data
@Accessors(chain = true)
public class Element {
    /**
     * The id of the element.
     */
    private long id;

    /**
     * Data cache.
     */
    private ByteBuffer buffer;

    /**
     * The constructor of Element.
     *
     * @param id The id of the element.
     * @param buffer Stored data
     */
    public Element(long id, ByteBuffer buffer) {
        this.id = id;
        this.buffer = buffer;
    }

    /**
     * Default constructor for Element.
     */
    public Element() {}

    @Override
    public String toString() {
        return super.toString();
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        Element otherElement = (Element) other;
        return id == otherElement.id;
    }

    @Override
    public int hashCode() {
        return Objects.hash(id);
    }
}
