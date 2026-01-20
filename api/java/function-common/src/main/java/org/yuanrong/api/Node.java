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

package org.yuanrong.api;

import lombok.Data;

import java.util.List;
import java.util.Map;

/**
 * The node information.
 *
 * @since 2025/07/28
 */
@Data
public class Node {
    private String id;
    private boolean alive;
    private Map<String, Float> resources;
    private Map<String, List<String>> labels;

    public Node() {}

    /**
     * Init Node with id, alive, resources and labels.
     *
     * @param id node id.
     * @param alive whether this node is alive.
     * @param resources resources of this node.
     * @param labels labels of this node.
     */
    public Node(String id, boolean alive, Map<String, Float> resources, Map<String, List<String>> labels) {
        this.id = id;
        this.alive = alive;
        this.resources = resources;
        this.labels = labels;
    }
}
