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

import org.yuanrong.affinity.Affinity;
import org.yuanrong.affinity.AffinityKind;
import org.yuanrong.affinity.AffinityScope;
import org.yuanrong.affinity.AffinityType;
import org.yuanrong.affinity.LabelOperator;
import org.yuanrong.runtime.util.Constants;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.google.gson.Gson;
import com.google.gson.JsonSyntaxException;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * InvokeOptions
 *
 * @since 2023 /11/30
 */
@Data
@Accessors(chain = true)
public class InvokeOptions {
    private static final Gson GSON = new Gson();

    private int cpu = 500;
    private int memory = 500;
    private Map<String, Float> customResources;
    private Map<String, String> customExtensions;
    private Map<String, String> podLabels;
    private List<String> labels;
    private Map<String, String> affinity;
    private int retryTimes = 0;
    private int priority = 0;
    private int instancePriority = 0;
    private long scheduleTimeoutMs = 30000L;
    private int recoverRetryTimes = 0;
    private List<Affinity> scheduleAffinities;
    private boolean preferredPriority = true;
    private boolean requiredPriority = false;
    private boolean preferredAntiOtherLabels = false;
    private boolean needOrder = false;
    private boolean preemptedAllowed = false;
    private String groupName;
    private Map<String, String> envVars;
    private String traceId;
    private Map<String, String> aliasParams;

    @JsonIgnore
    private Map<String, String> createOptions;

    /**
     * Instantiates new InvokeOptions.
     */
    public InvokeOptions() {
        this.customResources = new HashMap<>();
        this.customExtensions = new HashMap<>();
        this.createOptions = new HashMap<>();
        this.podLabels = new HashMap<>();
        this.labels = new ArrayList<>();
        this.affinity = new HashMap<>();
        this.scheduleAffinities = new ArrayList<>();
        this.envVars = new HashMap<>();
        this.aliasParams = new HashMap<>();
    }

    /**
     * Instantiates new InvokeOptions.
     *
     * @param opts the invoke options
     */
    public InvokeOptions(InvokeOptions opts) {
        this.cpu = opts.getCpu();
        this.memory = opts.getMemory();
        this.customResources = new HashMap<>(opts.getCustomResources());
        this.customExtensions = new HashMap<>(opts.getCustomExtensions());
        this.podLabels = new HashMap<>(opts.getPodLabels());
        this.labels = new ArrayList<>(opts.getLabels());
        this.affinity = new HashMap<>(opts.getAffinity());
        this.retryTimes = opts.getRetryTimes();
        this.priority = opts.getPriority();
        this.instancePriority = opts.getInstancePriority();
        this.scheduleTimeoutMs = opts.getScheduleTimeoutMs();
        this.recoverRetryTimes = opts.getRecoverRetryTimes();
        this.scheduleAffinities = new ArrayList<>(opts.getScheduleAffinities());
        this.preferredPriority = opts.isPreferredPriority();
        this.requiredPriority = opts.isRequiredPriority();
        this.preferredAntiOtherLabels = opts.isPreferredAntiOtherLabels();
        this.needOrder = opts.isNeedOrder();
        this.preemptedAllowed = opts.isPreemptedAllowed();
        this.groupName = opts.getGroupName();
        this.envVars = new HashMap<>(opts.getEnvVars());
        this.createOptions = new HashMap<>(opts.getCreateOptions());
        this.traceId = opts.getTraceId();
        this.aliasParams = opts.getAliasParams();
    }

    private class AffinityMsg {
        private List<Affinity> scheduleAffinities = new ArrayList<Affinity>();

        private boolean preferredPriority = true;

        private boolean requiredPriority = false;

        private AffinityMsg(boolean preferredPriority, boolean requiredPriority, List<Affinity> scheduleAffinities) {
            this.scheduleAffinities = scheduleAffinities;
            this.preferredPriority = preferredPriority;
            this.requiredPriority = requiredPriority;
        }
    }

    /**
     * getCreateOptions
     *
     * @return map
     */
    public Map<String, String> getCreateOptions() {
        return this.createOptions;
    }

    /**
     * getPodLabels
     *
     * @return map
     */
    public Map<String, String> getPodLabels() {
        return this.podLabels;
    }

    /**
     * Sets custom extensions.
     *
     * @param extension a map containing the custom extensions to be set.
     */
    public void setCustomExtensions(Map<String, String> extension) {
        this.customExtensions.putAll(extension);
        if (this.customExtensions.containsKey(Constants.POST_START_EXEC)) {
            this.createOptions.put(Constants.POST_START_EXEC, extension.get(Constants.POST_START_EXEC));
            this.customExtensions.remove(Constants.POST_START_EXEC);
        }
    }

    /**
     * addCustomExtensions
     *
     * @param key the customExtensions key
     * @param value the customExtensions value
     */
    public void addCustomExtensions(String key, String value) {
        if (Constants.POST_START_EXEC.equals(key)) {
            this.createOptions.put(key, value);
            return;
        }
        this.customExtensions.put(key, value);
    }

    /**
     * set affinityMsgToJsonStr
     *
     * @return String
     */
    public String affinityMsgToJsonStr() {
        AffinityMsg affinityMsg = new AffinityMsg(this.preferredPriority, this.requiredPriority,
            this.scheduleAffinities);
        return GSON.toJson(affinityMsg);
    }

    /**
     * set preferredPriority set scheduleAffinities
     *
     * @param jsonString the jsonString
     * @throws JsonSyntaxException If an error occurs during deserialization
     */
    public void parserAffinityMsgFromJsonStr(String jsonString) throws JsonSyntaxException {
        if (jsonString.isEmpty()) {
            return;
        }

        AffinityMsg affinityMsg = GSON.fromJson(jsonString, AffinityMsg.class);
        this.preferredPriority = affinityMsg.preferredPriority;
        this.scheduleAffinities = affinityMsg.scheduleAffinities;
    }

    /**
     * set retryTimes.
     *
     * @param retryTimes the retryTimes.
     */
    public void setRetryTimes(int retryTimes) {
        if (retryTimes < 0) {
            throw new IllegalArgumentException("retryTimes must be non-negative.");
        }
        this.retryTimes = retryTimes;
    }

    /**
     * set priority.
     *
     * @param priority the retryTimes.
     */
    public void setPriority(int priority) {
        if (priority < 0) {
            throw new IllegalArgumentException("priority must be non-negative.");
        }
        this.priority = priority;
    }

    /**
     * Returns the Builder object of InvokeOptions class.
     *
     * @return Builder
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of InvokeOptions class
     */
    public static class Builder {
        private InvokeOptions options;

        /**
         * Instantiates new InvokeOptions.
         */
        public Builder() {
            options = new InvokeOptions();
        }

        /**
         * set the cpu
         *
         * @param cpu the cpu num
         * @return InvokeOptions Builder class object.
         */
        public Builder cpu(int cpu) {
            options.setCpu(cpu);
            return this;
        }

        /**
         * set the memory
         *
         * @param mem the memory num
         * @return InvokeOptions Builder class object.
         */
        public Builder memory(int mem) {
            options.setMemory(mem);
            return this;
        }

        /**
         * set the recoverRetryTimes
         *
         * @param recoverRetryTimes the recoverRetryTimes
         * @return InvokeOptions Builder class object.
         */
        public Builder recoverRetryTimes(int recoverRetryTimes) {
            options.setRecoverRetryTimes(recoverRetryTimes);
            return this;
        }

        /**
         * set the customResources
         *
         * @param key the customResources key
         * @param val the customResources value
         * @return InvokeOptions Builder class object.
         */
        public Builder addCustomResource(String key, Float val) {
            options.customResources.put(key, val);
            return this;
        }

        /**
         * set the customExtensions
         *
         * @param key the customExtensions key
         * @param val the customExtensions value
         * @return InvokeOptions Builder class object.
         */
        public Builder addCustomExtension(String key, String val) {
            if (Constants.POST_START_EXEC.equals(key)) {
                options.createOptions.put(key, val);
                return this;
            }
            options.customExtensions.put(key, val);
            return this;
        }

        /**
         * set the podLabels
         *
         * @param key the podLabels key
         * @param val the podLabels value
         * @return InvokeOptions Builder class object.
         */
        public Builder addPodLabel(String key, String val) {
            options.podLabels.put(key, val);
            return this;
        }

        /**
         * set the aliasParams
         *
         * @param key the aliasParams key
         * @param val the aliasParams value
         * @return InvokeOptions Builder class object.
         */
        public Builder addAliasParam(String key, String val) {
            options.aliasParams.put(key, val);
            return this;
        }

        /**
         * set the labels
         *
         * @param val the label
         * @return InvokeOptions Builder class object.
         */

        public Builder addLabel(String val) {
            options.labels.add(val);
            return this;
        }

        /**
         * set the scheduleAffinities with AffinityKind.INSTANCE
         *
         * @param type the affinity type
         * @param operators the affinity operators
         * @return InvokeOptions Builder class object.
         */
        public Builder addInstanceAffinity(AffinityType type, List<LabelOperator> operators) {
            options.scheduleAffinities.add(new Affinity(AffinityKind.INSTANCE, type, operators));
            return this;
        }

        /**
         * set the scheduleAffinities with AffinityKind.INSTANCE and affinityScope
         *
         * @param type the affinity type
         * @param operators the affinity operators
         * @param affinityScope the affinity scope
         * @return InvokeOptions Builder class object.
         */
        public Builder addInstanceAffinity(AffinityType type, List<LabelOperator> operators,
            AffinityScope affinityScope) {
            options.scheduleAffinities.add(new Affinity(AffinityKind.INSTANCE, type, operators, affinityScope));
            return this;
        }

        /**
         * set the scheduleAffinities with AffinityKind.RESOURCE
         *
         * @param type the affinity type
         * @param operators the affinity operators
         * @return InvokeOptions Builder class object.
         */
        public Builder addResourceAffinity(AffinityType type, List<LabelOperator> operators) {
            options.scheduleAffinities.add(new Affinity(AffinityKind.RESOURCE, type, operators));
            return this;
        }

        /**
         * set the scheduleAffinities
         *
         * @param affinity the affinity
         * @return InvokeOptions Builder class object.
         */
        public Builder addScheduleAffinity(Affinity affinity) {
            options.scheduleAffinities.add(affinity);
            return this;
        }

        /**
         * set the scheduleAffinities
         *
         * @param affinities the affinities list
         * @return InvokeOptions Builder class object.
         */
        public Builder scheduleAffinity(List<Affinity> affinities) {
            options.scheduleAffinities = affinities;
            return this;
        }

        /**
         * set the preferredPriority
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder preferredPriority(boolean val) {
            options.preferredPriority = val;
            return this;
        }

        /**
         * set the requiredPriority
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder requiredPriority(boolean val) {
            options.requiredPriority = val;
            return this;
        }

        /**
         * set the preferredAntiOtherLabels
         *
         * @param isAntiOthers the isAntiOthers
         * @return InvokeOptions Builder class object.
         */
        public Builder preferredAntiOtherLabels(boolean isAntiOthers) {
            options.preferredAntiOtherLabels = isAntiOthers;
            return this;
        }

        /**
         * set the needOrder
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder needOrder(boolean val) {
            options.needOrder = val;
            return this;
        }

        /**
         * set the scheduleTimeoutMs
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder scheduleTimeoutMs(long val) {
            options.scheduleTimeoutMs = val;
            return this;
        }

        /**
         * set the preemptedAllowed
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder preemptedAllowed(boolean val) {
            options.preemptedAllowed = val;
            return this;
        }

        /**
         * set the priority
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder priority(int val) {
            if (val < 0) {
                throw new IllegalArgumentException("priority must be non-negative.");
            }
            options.priority = val;
            return this;
        }

        /**
         * set the instancePriority
         *
         * @param val the val
         * @return InvokeOptions Builder class object.
         */
        public Builder instancePriority(int val) {
            options.instancePriority = val;
            return this;
        }

        /**
         * set the groupname
         *
         * @param inputGroupName inputGroupName
         * @return InvokeOptions Builder class object.
         */
        public Builder groupName(String inputGroupName) {
            options.groupName = inputGroupName;
            return this;
        }

        /**
         * set the traceId
         *
         * @param inputTraceId inputTraceId
         * @return InvokeOptions Builder class object.
         */
        public Builder traceId(String inputTraceId) {
            options.traceId = inputTraceId;
            return this;
        }

        /**
         * set the customResources
         *
         * @param customResources the customResources map
         * @return InvokeOptions Builder class object.
         */
        public Builder customResources(Map<String, Float> customResources) {
            options.customResources.putAll(customResources);
            return this;
        }

        /**
         * set the customExtensions
         *
         * @param customExtensions the customExtensions map
         * @return InvokeOptions Builder class object.
         */
        public Builder customExtensions(Map<String, String> customExtensions) {
            options.customExtensions.putAll(customExtensions);
            String postStartValue = customExtensions.get(Constants.POST_START_EXEC);
            if (postStartValue != null) {
                options.createOptions.put(Constants.POST_START_EXEC, postStartValue);
                options.customExtensions.remove(Constants.POST_START_EXEC);
            }
            return this;
        }

        /**
         * set the podLabels
         *
         * @param podLabels the podLabels map
         * @return InvokeOptions Builder class object.
         */
        public Builder podLabels(Map<String, String> podLabels) {
            options.podLabels.putAll(podLabels);
            return this;
        }

        /**
         * set the aliasParams
         *
         * @param aliasParams the aliasParams map
         * @return InvokeOptions Builder class object.
         */
        public Builder aliasParams(Map<String, String> aliasParams) {
            options.aliasParams.putAll(aliasParams);
            return this;
        }

        /**
         * Builds the InvokeOptions object.
         *
         * @return InvokeOptions class object.
         */
        public InvokeOptions build() {
            return options;
        }
    }
}
