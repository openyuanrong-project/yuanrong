
/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
import org.yuanrong.affinity.OperatorType;
import org.yuanrong.runtime.util.Constants;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.modules.junit4.PowerMockRunner;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore("javax.management.*")
public class TestInvokeOptions {
    @Test
    public void testPreferredAntiOtherLabelsIsFalse() {
        final boolean expectedValue = false;
        final InvokeOptions options = new InvokeOptions.Builder()
                .preferredAntiOtherLabels(expectedValue)
                .build();
        options.getCpu();
        Assert.assertEquals(expectedValue, options.isPreferredAntiOtherLabels());
    }

    @Test
    public void testPreferredPriorityIsFalse() {
        boolean expectedValue = false;
        InvokeOptions options = new InvokeOptions.Builder().preferredPriority(expectedValue).build();
        options.getGroupName();
        options.getAffinity();
        Assert.assertEquals(expectedValue, options.isPreferredPriority());
    }

    @Test
    public void testNeedOrderIsTrue() {
        boolean expectedValue = true;
        InvokeOptions options = new InvokeOptions.Builder()
            .needOrder(expectedValue).build();
        Assert.assertEquals(expectedValue, options.isNeedOrder());
    }

    @Test
    public void testCpuNums() {
        int expectedValue = 1;
        InvokeOptions options = new InvokeOptions.Builder()
            .cpu(expectedValue)
            .build();
        Assert.assertEquals(expectedValue, options.getCpu());
    }

    @Test
    public void testMemoryNums() {
        int expectedValue = 1;
        InvokeOptions options = new InvokeOptions.Builder()
            .memory(expectedValue)
            .build();
        Assert.assertEquals(expectedValue, options.getMemory());
    }

    @Test
    public void testPriority() {
        int expectedValue = 22;
        InvokeOptions options = new InvokeOptions.Builder()
            .priority(22)
            .build();
        Assert.assertEquals(expectedValue, options.getPriority());
    }

    @Test
    public void testInstancePriority() {
        int expectedValue = 22;
        InvokeOptions options = new InvokeOptions.Builder()
            .instancePriority(22)
            .build();
        Assert.assertEquals(expectedValue, options.getInstancePriority());
    }

    @Test
    public void testGroupName() {
        String inputGroupName = "testGroupName";
        InvokeOptions options = new InvokeOptions.Builder()
            .groupName(inputGroupName)
            .build();
        Assert.assertEquals(inputGroupName, options.getGroupName());
    }

    @Test
    public void testTraceId() {
        String inputTraceId = "testTraceId";
        InvokeOptions options = new InvokeOptions.Builder()
            .traceId(inputTraceId)
            .build();
        Assert.assertEquals(inputTraceId, options.getTraceId());
    }

    @Test
    public void testPreemptedAllowed() {
        boolean expectedValue = true;
        InvokeOptions options = new InvokeOptions.Builder()
            .preemptedAllowed(true)
            .build();
        Assert.assertEquals(expectedValue, options.isPreemptedAllowed());
    }

    @Test
    public void testScheduleTimeoutMs() {
        long expectedValue = 22222L;
        InvokeOptions options = new InvokeOptions.Builder()
            .scheduleTimeoutMs(22222L)
            .build();
        Assert.assertEquals(expectedValue, options.getScheduleTimeoutMs());
    }

    @Test
    public void testCustomResourcesSize() {
        String expectedMapKey = "gpu-time";
        Float expectedMapValue = 1.5f;
        int expectedSize = 2;
        HashMap<String, Float> customResources = new HashMap<>();
        customResources.put("accelerator", 0.5f);
        InvokeOptions options = new InvokeOptions.Builder()
            .customResources(customResources)
            .addCustomResource(expectedMapKey,expectedMapValue)
            .build();
        Assert.assertEquals(expectedSize, options.getCustomResources().size());
    }

    @Test
    public void testCustomExtensionsSize() {
        String expectedMapKey = "test-extension";
        String expectedMapValue = "true";
        HashMap<String, String> testCustomExtensions = new HashMap<>();
        testCustomExtensions.put(expectedMapKey, expectedMapValue);
        testCustomExtensions.put(Constants.POST_START_EXEC, "false");
        InvokeOptions options = new InvokeOptions.Builder().customExtensions(testCustomExtensions)
            .addCustomExtension(expectedMapKey, expectedMapValue)
            .addCustomExtension(Constants.POST_START_EXEC, "true")
            .build();
        options.setCustomExtensions(testCustomExtensions);
        options.addCustomExtensions(expectedMapKey, expectedMapValue);
        options.addCustomExtensions(Constants.POST_START_EXEC, "true");
        Assert.assertEquals(1, options.getCustomExtensions().size());
    }

    @Test
    public void testPodLabelsSize() {
        String expectedMapKey = "pod-a";
        String expectedMapValue = "test-pod";
        HashMap<String, String> testPodLabels = new HashMap<>();
        testPodLabels.put(expectedMapKey, expectedMapValue);
        InvokeOptions options = new InvokeOptions.Builder().podLabels(testPodLabels)
            .addPodLabel(expectedMapKey, expectedMapValue)
            .build();
        Assert.assertEquals(1, options.getPodLabels().size());
    }

    @Test
    public void testAliasParamsSize() {
        String key = "pod-a";
        String value = "test-pod";
        HashMap<String, String> aliasParams = new HashMap<>();
        aliasParams.put(key, value);
        InvokeOptions options = new InvokeOptions.Builder().aliasParams(aliasParams)
                .addAliasParam(key, value)
                .build();
        Assert.assertEquals(1, options.getAliasParams().size());
    }

    @Test
    public void testLabelsSize() {
        String expectedValue = "lable-a";
        InvokeOptions options = new InvokeOptions.Builder().addLabel(expectedValue).build();
        Assert.assertEquals(1, options.getLabels().size());
    }

    @Test
    public void testRecoverRetryTimes() {
        int expectedValue = 1;
        InvokeOptions options = new InvokeOptions.Builder().recoverRetryTimes(expectedValue).build();
        Assert.assertEquals(expectedValue, options.getRecoverRetryTimes());
    }

    @Test
    public void testCreateOptionsSize() {
        String expectedMapKey = "option-a";
        String expectedMapValue = "test-option";
        HashMap<String, String> testCreateOptions = new HashMap<>();
        testCreateOptions.put(expectedMapKey, expectedMapValue);
        InvokeOptions options = InvokeOptions.builder().build();
        options.setCreateOptions(testCreateOptions);
        Assert.assertEquals(1, options.getCreateOptions().size());
    }

    @Test
    public void testScheduleAffinitiesSize() {
        LabelOperator labelOperator = new LabelOperator(OperatorType.LABEL_IN, "test-label");
        ArrayList<LabelOperator> testOperatorsList = new ArrayList<>();
        testOperatorsList.add(labelOperator);
        Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, testOperatorsList);
        ArrayList<Affinity> testAffinityList = new ArrayList<>();
        testAffinityList.add(affinity);
        InvokeOptions options = new InvokeOptions.Builder().scheduleAffinity(testAffinityList)
            .addInstanceAffinity(AffinityType.PREFERRED, testOperatorsList)
            .addInstanceAffinity(AffinityType.PREFERRED, testOperatorsList, AffinityScope.NODE)
            .addResourceAffinity(AffinityType.PREFERRED, testOperatorsList)
            .addScheduleAffinity(affinity)
            .build();
        options.parserAffinityMsgFromJsonStr("");
        options.parserAffinityMsgFromJsonStr(options.affinityMsgToJsonStr());
        Assert.assertEquals(5, options.getScheduleAffinities().size());

        InvokeOptions newOptions = new InvokeOptions(options);
        List<Affinity> newList = newOptions.getScheduleAffinities();
        List<Affinity> oldList = options.getScheduleAffinities();
        Assert.assertTrue(newList.size() == oldList.size());
        for (int i = 0; i < newList.size(); i++) {
            Assert.assertTrue(newList.get(i).getAffinityKind().compareTo(oldList.get(i).getAffinityKind()) == 0);
            Assert.assertTrue(newList.get(i).getAffinityType().compareTo(oldList.get(i).getAffinityType()) == 0);
            List<LabelOperator> newLabelOperators = newList.get(i).getLabelOperators();
            List<LabelOperator> oldLabelOperators = oldList.get(i).getLabelOperators();
            Assert.assertTrue(newLabelOperators.size() == oldLabelOperators.size());
            for (int j = 0; j < newLabelOperators.size(); j++) {
                Assert.assertTrue(newLabelOperators.get(j).getOperatorType().compareTo(oldLabelOperators.get(j).getOperatorType()) == 0);
                Assert.assertTrue(newLabelOperators.get(j).getKey() == oldLabelOperators.get(j).getKey());
                Assert.assertTrue(newLabelOperators.get(j).getValues().size() == oldLabelOperators.get(j).getValues().size());
            }
        }
    }

    @Test
    public void testData() {
        InvokeOptions options = new InvokeOptions.Builder().build();
        options.setCpu(600);
        options.setMemory(600);
        options.setCustomResources(new HashMap<String, Float>(){{put("npu", 10.00f);}});
        options.setPodLabels(new HashMap<String, String>(){{put("pod1", "label1");}});
        options.setLabels(new ArrayList<String>(){{add("label1");}});
        options.setAffinity(new HashMap<>());
        options.setPreferredPriority(true);
        options.setPreferredAntiOtherLabels(true);
        options.setGroupName("groupName");
        options.setRetryTimes(1);
        options.setPriority(2);
        options.setPreemptedAllowed(true);
        options.setScheduleTimeoutMs(22222L);
        options.setScheduleAffinities(new ArrayList<>());
        options.getRetryTimes();
        options.getPriority();
        options.isPreemptedAllowed();
        options.getScheduleTimeoutMs();
        InvokeOptions testOptions = new InvokeOptions.Builder().build();
        options.canEqual(testOptions);
        options.toString();
        options.hashCode();
        Assert.assertFalse(options.equals(testOptions));

        InvokeOptions newOptions = new InvokeOptions(options);
        Assert.assertTrue(newOptions.getCpu() == options.getCpu());
        Assert.assertTrue(newOptions.getMemory() == options.getMemory());
        Assert.assertTrue(newOptions.getCustomResources().size() == options.getCustomResources().size());
        Assert.assertTrue(newOptions.getLabels().size() == options.getLabels().size());
        Assert.assertTrue(newOptions.getScheduleTimeoutMs() == options.getScheduleTimeoutMs());
        Assert.assertTrue(newOptions.isPreemptedAllowed() == options.isPreemptedAllowed());
        Assert.assertTrue(newOptions.getGroupName() == options.getGroupName());
        Assert.assertTrue(newOptions.getPodLabels().size() == options.getPodLabels().size());
    }
}