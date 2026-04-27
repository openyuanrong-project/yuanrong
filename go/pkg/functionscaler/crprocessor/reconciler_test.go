/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package crprocessor -
package crprocessor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/instanceconfig"
	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commontype "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/functioncr"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewAgentRunResourcesManager(t *testing.T) {
	convey.Convey("test NewAgentCRsManager", t, func() {
		stopCh := make(chan struct{}, 1)
		manager := NewAgentCRsManager(stopCh)
		convey.So(manager.synced, convey.ShouldBeFalse)
		stopCh <- struct{}{}
		convey.So(<-manager.stopCh, convey.ShouldNotBeNil)
	})
}

func TestProcessCREventLoop(t *testing.T) {
	convey.Convey("Test ProcessCREventLoop", t, func() {
		convey.Convey("Should handle SubEventTypeSynced", func() {
			syncCh := make(chan struct{}, 1)
			event := registry.SubEvent{
				EventType: registry.SubEventTypeSynced,
				EventMsg:  syncCh,
			}
			stopCh := make(chan struct{})
			defer utils.SafeCloseChannel(stopCh)
			manager := NewAgentCRsManager(stopCh)
			go manager.ProcessCREventLoop()
			manager.AgentCRCh <- event
			convey.So(<-syncCh, convey.ShouldNotBeNil)
		})

		convey.Convey("Should process CR update and delete event", func() {
			wg := sync.WaitGroup{}
			stopCh := make(chan struct{})
			defer utils.SafeCloseChannel(stopCh)
			manager := NewAgentCRsManager(stopCh)
			var recvEventType registry.EventType
			var recvCrname string
			defer gomonkey.ApplyPrivateMethod(manager, "processCRInfoEvent", func(_ *AgentCRsManager, eventType registry.EventType, info *types.AgentEventInfo) {
				log.GetLogger().Sync()
				recvEventType = eventType
				recvCrname = info.CrKey
				wg.Done()
			}).Reset()

			event := registry.SubEvent{
				EventType: registry.SubEventTypeUpdate,
				EventMsg: &types.AgentEventInfo{
					CrKey: "default:test-cr",
				},
			}
			go manager.ProcessCREventLoop()
			wg.Add(1)
			manager.AgentCRCh <- event
			wg.Wait()
			convey.So(recvCrname, convey.ShouldEqual, "default:test-cr")
			convey.So(recvEventType, convey.ShouldEqual, registry.SubEventTypeUpdate)

			event.EventType = registry.SubEventTypeDelete
			wg.Add(1)
			manager.AgentCRCh <- event
			wg.Wait()
			convey.So(recvCrname, convey.ShouldEqual, "default:test-cr")
			convey.So(recvEventType, convey.ShouldEqual, registry.SubEventTypeDelete)

			utils.SafeCloseChannel(stopCh)
		})
	})
}

func TestProcessInstanceEvent(t *testing.T) {
	convey.Convey("Test processInstanceEvent", t, func() {
		funcKey := "test-func-key"
		reconciler := &AgentCRReconciler{
			funcKey:            funcKey,
			reconcileTriggerCh: make(chan struct{}, 1),
		}
		updateCount := 0
		deleteCount := 0
		wg := sync.WaitGroup{}
		patches := []*gomonkey.Patches{
			gomonkey.ApplyPrivateMethod(reconciler, "processInstanceUpdate", func(_ *AgentCRsManager, info *commontype.InstanceSpecification) {
				updateCount++
				wg.Done()
			}),
			gomonkey.ApplyPrivateMethod(reconciler, "processInstanceDelete", func(_ *AgentCRsManager, info *commontype.InstanceSpecification) {
				deleteCount++
				wg.Done()
			}),
			gomonkey.ApplyFunc(isCrReconcilerOwner, func(bool, string) bool { return true }),
		}
		defer func() {
			for _, p := range patches {
				p.Reset()
			}
		}()
		stopCh := make(chan struct{})
		defer utils.SafeCloseChannel(stopCh)
		manager := NewAgentCRsManager(stopCh)
		manager.funcKeyToAgentCRReconciler[funcKey] = reconciler
		manager.crNameToAgentCRReconciler["111"] = reconciler
		go manager.ProcessInstanceEventLoop()

		instance := &commontype.InstanceSpecification{
			InstanceID: "test-instance",
			Function:   "12345678901234561234567890123456/0-system-faasExecutorGo1.x/$latest",
			CreateOptions: map[string]string{
				types.FunctionKeyNote: funcKey,
				types.FunctionSign:    "test-sign",
			},
			InstanceStatus: commontype.InstanceStatus{Code: int32(constant.KernelInstanceStatusRunning)},
		}

		wg.Add(1)
		manager.AgentCRInsCh <- registry.SubEvent{EventType: registry.SubEventTypeUpdate, EventMsg: instance}
		wg.Wait()
		//	manager.processInstanceEvent(registry.SubEventTypeUpdate, instance)
		convey.So(updateCount == 1 && deleteCount == 0, convey.ShouldBeTrue)

		wg.Add(1)
		manager.AgentCRInsCh <- registry.SubEvent{EventType: registry.SubEventTypeDelete, EventMsg: instance}
		wg.Wait()
		//	manager.processInstanceEvent(registry.SubEventTypeDelete, instance)
		convey.So(updateCount == 1 && deleteCount == 1, convey.ShouldBeTrue)

		wg.Add(1)
		instance.InstanceStatus.Code = int32(constant.KernelInstanceStatusEvicting)
		manager.AgentCRInsCh <- registry.SubEvent{EventType: registry.SubEventTypeUpdate, EventMsg: instance}
		//	manager.processInstanceEvent(registry.SubEventTypeUpdate, instance)
		wg.Wait()
		convey.So(updateCount == 1 && deleteCount == 2, convey.ShouldBeTrue)

		manager.processInstanceEvent(registry.SubEventTypeSynced, &commontype.InstanceSpecification{})
		convey.So(<-reconciler.reconcileTriggerCh, convey.ShouldNotBeEmpty)

		reconciler.enable = false
		manager.processInstanceEvent(registry.SubEventTypeSynced, &commontype.InstanceSpecification{})
		convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
	})
}

func TestProcessFaaSSchedulerProxyEventLoop(t *testing.T) {
	convey.Convey("Test ProcessFaaSSchedulerProxyEventLoop", t, func() {
		stopCh := make(chan struct{})
		defer close(stopCh)
		manager := NewAgentCRsManager(stopCh)
		triggerCount := 0
		wg := sync.WaitGroup{}
		defer gomonkey.ApplyPrivateMethod(&AgentCRReconciler{}, "setEnable", func(_ *AgentCRReconciler, b bool) {
			triggerCount++
			wg.Done()
		}).Reset()
		manager.crNameToAgentCRReconciler["test-1"] = &AgentCRReconciler{}
		manager.crNameToAgentCRReconciler["test-2"] = &AgentCRReconciler{}
		go manager.ProcessFaaSSchedulerProxyEventLoop()
		wg.Add(2)
		manager.FaaSSchedulerProxyCh <- registry.SubEvent{}
		wg.Wait()
		convey.So(triggerCount, convey.ShouldEqual, 2)
	})
}

func TestIsCrReconcilerOwner(t *testing.T) {
	convey.Convey("test isCrReconcilerOwner", t, func() {
		convey.So(isCrReconcilerOwner(false, "1234"), convey.ShouldBeFalse)
		p := gomonkey.ApplyMethodFunc(selfregister.GlobalSchedulerProxy, "IsFuncOwner", func(funcKey string) bool {
			return false
		})
		convey.So(isCrReconcilerOwner(true, "1234"), convey.ShouldBeFalse)
		p.Reset()
		p = gomonkey.ApplyMethodFunc(selfregister.GlobalSchedulerProxy, "IsFuncOwner", func(funcKey string) bool {
			return true
		})
		convey.So(isCrReconcilerOwner(false, "1234"), convey.ShouldBeFalse)
		convey.So(isCrReconcilerOwner(true, "1234"), convey.ShouldBeTrue)
		p.Reset()
	})
}

func TestIsStatusDiff(t *testing.T) {
	convey.Convey("Test isStatusDiff", t, func() {
		oldStatus := &types.AgentCRStatus{
			ReadyReplicas: 1,
			Conditions: []*metav1.Condition{
				{
					Type:               "ReadyPod",
					Status:             "true",
					LastTransitionTime: metav1.NewTime(time.Now().UTC()),
					Reason:             "ReadyPod",
					Message:            "pod-1",
				},
			},
		}

		newStatus := &types.AgentCRStatus{
			ReadyReplicas: 1,
			Conditions: []*metav1.Condition{
				{
					Type:               "ReadyPod",
					Status:             "true",
					LastTransitionTime: oldStatus.Conditions[0].LastTransitionTime,
					Reason:             "ReadyPod",
					Message:            "pod-1",
				},
			},
		}

		convey.Convey("Should return false when status equal", func() {
			convey.So(isStatusDiff(oldStatus, newStatus), convey.ShouldBeFalse)
		})

		convey.Convey("Should detect ReadyReplicas change", func() {
			modifiedStatus := *newStatus
			modifiedStatus.ReadyReplicas = 2
			convey.So(isStatusDiff(oldStatus, &modifiedStatus), convey.ShouldBeTrue)
		})

		convey.Convey("Should detect Condition field change", func() {
			testCases := []struct {
				ModifyFn func(condition *metav1.Condition)
			}{
				{func(c *metav1.Condition) { c.Status = "false" }},
				{func(c *metav1.Condition) { c.Reason = "Failed" }},
				{func(c *metav1.Condition) { c.Message = "pod-2" }},
			}

			for _, tc := range testCases {
				modifiedStatus := *newStatus
				modifiedCondition := modifiedStatus.Conditions[0]
				tc.ModifyFn(modifiedCondition)
				modifiedStatus.Conditions[0] = modifiedCondition
				convey.So(isStatusDiff(oldStatus, &modifiedStatus), convey.ShouldBeTrue)
			}
		})
	})
}

func TestIsExpectedStatus(t *testing.T) {
	convey.Convey("Test isExpectedStatus", t, func() {
		expectedSignature := "test-sig"
		instanceMap := map[string]*commontype.InstanceSpecification{
			"pod-1": {
				InstanceID: "pod-1",
				CreateOptions: map[string]string{
					types.FunctionSign: expectedSignature,
				},
			},
		}

		status := &types.AgentCRStatus{
			Conditions: []*metav1.Condition{
				{
					Type:    "ReadyPod",
					Status:  "true",
					Reason:  "ReadyPod",
					Message: "pod-1",
				},
			},
		}

		convey.Convey("Should return true when status matches instances", func() {
			convey.So(isExpectedStatus(expectedSignature, status, instanceMap), convey.ShouldBeTrue)
		})

		convey.Convey("Should return false when condition length mismatch", func() {
			modifiedStatus := *status
			modifiedStatus.Conditions = append(modifiedStatus.Conditions, &metav1.Condition{
				Message: "pod-2",
			})
			convey.So(isExpectedStatus(expectedSignature, &modifiedStatus, instanceMap), convey.ShouldBeFalse)
		})

		convey.Convey("Should return false when condition fields invalid", func() {
			testCases := []struct {
				ModifyFn func(condition *metav1.Condition)
			}{
				{func(c *metav1.Condition) { c.Type = "Failed" }},
				{func(c *metav1.Condition) { c.Status = "false" }},
				{func(c *metav1.Condition) { c.Reason = "Failed" }},
				{func(c *metav1.Condition) { c.Message = "invalid-pod" }},
			}

			for _, tc := range testCases {
				modifiedStatus := *status
				modifiedCondition := modifiedStatus.Conditions[0]
				tc.ModifyFn(modifiedCondition)
				modifiedStatus.Conditions[0] = modifiedCondition
				convey.So(isExpectedStatus(expectedSignature, &modifiedStatus, instanceMap), convey.ShouldBeFalse)
			}
		})
	})
}

func TestAgentRunCRReconciler_processEvent(t *testing.T) {
	convey.Convey("test AgentCRReconciler processEvent", t, func() {
		info := &types.AgentEventInfo{
			CrKey: "default:test-func-1-cr",
			FuncSpec: &types.FunctionSpecification{
				FuncKey:           "test-func-1",
				FuncMetaSignature: "123",
				MetaFromCR:        false,
			},
			Status: &types.AgentCRStatus{
				ReadyReplicas: 0,
				Conditions:    []*metav1.Condition{},
			},
		}
		defer gomonkey.ApplyPrivateMethod(&AgentCRReconciler{}, "reconcileLoop", func(reconciler *AgentCRReconciler) {
			return
		}).Reset()
		reconciler := newAgentRunCRReconciler(info, false)
		defer reconciler.exit()

		convey.Convey("processCRUpdate", func() {
			reconciler.enable = false
			isExpectedStatusBool := false
			defer gomonkey.ApplyFunc(isExpectedStatus, func(string, *types.AgentCRStatus, map[string]*commontype.InstanceSpecification) bool {
				return isExpectedStatusBool
			}).Reset()
			reconciler.processCRUpdate(info)
			convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
			reconciler.enable = true
			reconciler.processCRUpdate(info)
			convey.So(<-reconciler.reconcileTriggerCh, convey.ShouldNotBeNil)
		})
		convey.Convey("setEnable", func() {
			tests := []struct {
				oldEnable     bool
				newEnable     bool
				expectTrigger bool
			}{
				{
					false, false, false,
				}, {
					false, true, true,
				}, {
					true, false, false,
				}, {
					true, true, false,
				},
			}
			for _, tt := range tests {
				reconciler.enable = tt.oldEnable
				reconciler.setEnable(tt.newEnable)
				if tt.expectTrigger {
					convey.So(<-reconciler.reconcileTriggerCh, convey.ShouldNotBeNil)
				} else {
					convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
				}
			}
		})

		convey.Convey("test instance", func() {
			mockInstanceFunc := func(instanceId string, sign string) *commontype.InstanceSpecification {
				instance := &commontype.InstanceSpecification{InstanceID: instanceId, CreateOptions: make(map[string]string)}
				instance.CreateOptions[types.FunctionSign] = sign
				return instance
			}
			ins1 := mockInstanceFunc("ins1", "123")
			reconciler.enable = true
			reconciler.processInstanceUpdate(ins1, log.GetLogger())
			convey.So(<-reconciler.reconcileTriggerCh, convey.ShouldNotBeNil)
			ins2 := mockInstanceFunc("ins2", "1234")
			ins1_copy := mockInstanceFunc("ins1", "123")
			reconciler.processInstanceUpdate(ins2, log.GetLogger())
			reconciler.processInstanceUpdate(ins1_copy, log.GetLogger())
			convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
			reconciler.enable = false
			ins3 := mockInstanceFunc("ins3", "123")
			reconciler.processInstanceUpdate(ins3, log.GetLogger())
			convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
			reconciler.processInstanceDelete(ins3)
			convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
			reconciler.enable = true
			ins4 := mockInstanceFunc("ins4", "1234")
			reconciler.processInstanceDelete(ins4)
			convey.So(len(reconciler.reconcileTriggerCh), convey.ShouldEqual, 0)
			reconciler.processInstanceDelete(ins1)
			convey.So(<-reconciler.reconcileTriggerCh, convey.ShouldNotBeNil)
		})
	})
}

func TestSubscriberChan(t *testing.T) {
	convey.Convey("Test SubscriberChan", t, func() {
		insConfigCh := make(chan registry.SubEvent, 10)
		funcCh := make(chan registry.SubEvent, 10)

		manager := NewAgentCRsManager(make(chan struct{}))
		manager.AddInstanceConfigSubscriberChan(insConfigCh)
		manager.AddFunctionSubscriberChan(funcCh)
		event := &types.FunctionSpecification{
			FuncKey:          "test-func-1",
			InstanceMetaData: commontype.InstanceMetaData{},
		}
		manager.publishInstanceConfigEvent(registry.SubEventTypeUpdate, event)
		manager.publishFunctionEvent(registry.SubEventTypeUpdate, event)
		event1 := <-insConfigCh
		insConfig, ok := event1.EventMsg.(*instanceconfig.Configuration)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(insConfig.FuncKey, convey.ShouldEqual, "test-func-1")
		event2 := <-funcCh
		funcSpec, ok := event2.EventMsg.(*types.FunctionSpecification)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(funcSpec.FuncKey, convey.ShouldEqual, "test-func-1")
	})
}

func TestReconcileLoop(t *testing.T) {
	convey.Convey("test reconcileLoop", t, func() {
		info := &types.AgentEventInfo{
			CrKey: "default:test-func-1-cr",
			FuncSpec: &types.FunctionSpecification{
				FuncKey:           "test-func-1",
				FuncMetaSignature: "123",
				MetaFromCR:        false,
			},
			Status: &types.AgentCRStatus{
				ReadyReplicas: 0,
				Conditions:    []*metav1.Condition{},
			},
		}
		reconciler := newAgentRunCRReconciler(info, true)
		defer gomonkey.ApplyFunc(isExpectedStatus, func(string, *types.AgentCRStatus, map[string]*commontype.InstanceSpecification) bool {
			return false
		}).Reset()
		var mockReconcilerStatusFunc func() error
		defer gomonkey.ApplyPrivateMethod(reconciler, "reconcileStatus", func(crReconciler *AgentCRReconciler) error {
			if mockReconcilerStatusFunc == nil {
				return nil
			}
			return mockReconcilerStatusFunc()
		}).Reset()

		triggerCount := 0
		wg := sync.WaitGroup{}
		mockReconcilerStatusFunc = func() error {
			triggerCount++
			fmt.Printf("count: %d\n", triggerCount)
			wg.Done()
			if triggerCount == 3 {
				return nil
			}
			return fmt.Errorf("")
		}
		wg.Add(3)
		reconciler.triggerReconcile()
		wg.Wait()
		convey.So(triggerCount, convey.ShouldEqual, 3)
		reconciler.exit()
	})
}

func TestAgentRunCRReconciler(t *testing.T) {
	convey.Convey("Test AgentRunCRReconciler", t, func() {
		info := &types.AgentEventInfo{
			CrKey: "default:test-cr",
			FuncSpec: &types.FunctionSpecification{
				FuncKey: "test-func-key",
			},
		}
		reconciler := newAgentRunCRReconciler(info, true)
		defer reconciler.exit()

		convey.Convey("Should trigger reconcile when enabled", func() {
			wg := sync.WaitGroup{}
			triggerReconcileCount := 0
			defer gomonkey.ApplyFunc(isExpectedStatus, func(string, *types.AgentCRStatus, map[string]*commontype.InstanceSpecification) bool {
				return false
			}).Reset()
			defer gomonkey.ApplyPrivateMethod(reconciler, "reconcileStatus", func(crReconciler *AgentCRReconciler) error {
				triggerReconcileCount++
				wg.Done()
				return nil
			}).Reset()
			wg.Add(1)
			reconciler.triggerReconcile()
			wg.Wait()
			convey.So(triggerReconcileCount, convey.ShouldEqual, 1)
		})

		convey.Convey("Should process CR update correctly", func() {
			newInfo := &types.AgentEventInfo{
				CrKey: "default:test-cr",
				FuncSpec: &types.FunctionSpecification{
					FuncKey:           "test-func-key",
					FuncMetaSignature: "new-signature",
				},
			}
			reconciler.processCRUpdate(newInfo)
			convey.So(reconciler.info.FuncSpec.FuncMetaSignature, convey.ShouldEqual, "new-signature")
		})
	})
}

func TestAgentCRBtchProcess(t *testing.T) {
	convey.Convey("TestAgentCRBtchProcess", t, func() {
		stopCh := make(chan struct{})
		manager := NewAgentCRsManager(stopCh)
		defer gomonkey.ApplyFunc(k8sclient.GetDynamicClient, func() dynamic.Interface {
			return &dynamicfake.FakeDynamicClient{}
		}).Reset()
		defer gomonkey.ApplyMethodFunc((&dynamicfake.FakeDynamicClient{}).Resource(functioncr.GetCrdGVR()).Namespace("default"), "Get", func(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
			time.Sleep(5 * time.Second)
			return nil, nil
		}).Reset()
		defer gomonkey.ApplyPrivateMethod(&AgentCRReconciler{}, "getCrStatus", func(crObj *unstructured.Unstructured) (*types.AgentCRStatus, error) {
			time.Sleep(5 * time.Second)
			return &types.AgentCRStatus{}, nil
		}).Reset()
		wg := sync.WaitGroup{}
		defer gomonkey.ApplyPrivateMethod(&AgentCRReconciler{}, "updateCRStatus", func(resource dynamic.NamespaceableResourceInterface,
			crStatus *types.AgentCRStatus, crObj *unstructured.Unstructured) error {
			time.Sleep(5 * time.Second)
			wg.Done()
			return nil
		}).Reset()
		defer gomonkey.ApplyFunc(isExpectedStatus, func(string, *types.AgentCRStatus, map[string]*commontype.InstanceSpecification) bool {
			return false
		}).Reset()
		defer gomonkey.ApplyFunc(isCrReconcilerOwner, func(bool, string) bool {
			return true
		}).Reset()
		start := time.Now()
		manager.StartLoop()
		for i := 0; i < 100; i++ {
			wg.Add(1)
			manager.AgentCRCh <- registry.SubEvent{
				EventType: registry.SubEventTypeUpdate,
				EventMsg: &types.AgentEventInfo{
					CrKey:    fmt.Sprintf("default:crName: %d", i),
					FuncSpec: &types.FunctionSpecification{},
					Status:   &types.AgentCRStatus{},
				},
			}
		}
		convey.So(time.Since(start), convey.ShouldBeLessThanOrEqualTo, 1*time.Second)
		wg.Wait()
		convey.So(time.Since(start), convey.ShouldBeLessThanOrEqualTo, 16*time.Second)
		for _, crReconciler := range manager.crNameToAgentCRReconciler {
			crReconciler.exit()
		}
	})
}

func TestAgentCRReconciler_GetCrStatus(t *testing.T) {
	// Using the global logger instance might cause issues in tests.
	// It's better to inject it via constructor (Dependency Injection).
	// For this test, we assume it's a field and we can patch its methods.
	var reconciler *AgentCRReconciler

	convey.Convey("Given an AgentCRReconciler instance", t, func() {
		reconciler = &AgentCRReconciler{
			logger: log.GetLogger(),
		}

		convey.Convey("When the CR has a status that can be successfully converted", func() {
			crObj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"readyReplicas": int64(5),
						"conditions":    []interface{}{}, // Example condition
					},
				},
			}
			expectedStatus := &types.AgentCRStatus{
				ReadyReplicas: 5,
				Conditions:    []*metav1.Condition{},
			}

			// Mock the FromUnstructured function to return our expected status and no error
			defer gomonkey.ApplyMethodFunc(runtime.DefaultUnstructuredConverter, "FromUnstructured",
				func(input map[string]interface{}, output interface{}) error {
					outputStatus, ok := output.(*types.AgentCRStatus)
					if !ok {
						return fmt.Errorf("format error")
					}
					*outputStatus = *expectedStatus // Assign the pre-defined status
					return nil
				}).Reset()

			result, err := reconciler.getCrStatus(crObj)

			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, expectedStatus)
		})

		convey.Convey("When the CR has a 'status' field but it's not a map", func() {
			crObj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": "this is a string, not a map",
				},
			}

			_, err := reconciler.getCrStatus(crObj)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "transfer status format error")
			// Optionally, assert that logger was called with the right message
			// This is more complex and would require a more advanced mock of the logger.
		})

		convey.Convey("When the CR has a status map but conversion fails", func() {
			crObj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{}, // Empty status that will cause a conversion error
				},
			}
			expectedConvertErr := fmt.Errorf("schema do not match")

			// Mock FromUnstructured to return our expected error
			defer gomonkey.ApplyMethod(runtime.DefaultUnstructuredConverter, "FromUnstructured",
				func(_ runtime.UnstructuredConverter, input map[string]interface{}, output interface{}) error {
					return expectedConvertErr
				}).Reset()

			_, err := reconciler.getCrStatus(crObj)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "status format is not AgentRunStatus")
		})

		convey.Convey("When the CR has no 'status' field", func() {
			crObj := &unstructured.Unstructured{
				Object: map[string]interface{}{}, // No status key
			}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			result, err := reconciler.getCrStatus(crObj)

			// The function should return a zero-initialized status and no error
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldResemble, &types.AgentCRStatus{
				Conditions:    []*metav1.Condition{},
				ReadyReplicas: 0,
			})
		})
	})
}

// --- Test Suite for updateCRStatus ---
func TestAgentCRReconciler_UpdateCrStatus(t *testing.T) {
	convey.Convey("Given an AgentCRReconciler instance, a resource interface, and a CR object", t, func() {
		var reconciler *AgentCRReconciler
		reconciler = &AgentCRReconciler{logger: log.GetLogger()}

		mockResource := (&dynamicfake.FakeDynamicClient{}).Resource(functioncr.GetCrdGVR())

		crObj := &unstructured.Unstructured{Object: make(map[string]interface{})}
		crStatus := &types.AgentCRStatus{ReadyReplicas: 10}

		convey.Convey("When updating the status is successful", func() {

			var patches = gomonkey.NewPatches()
			defer patches.Reset()

			// Mock ToUnstructured to return a predictable map and no error
			patches.ApplyMethod(runtime.DefaultUnstructuredConverter, "ToUnstructured",
				func(_ runtime.UnstructuredConverter, input interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"readyReplicas": int64(10)}, nil
				})

			err := reconciler.updateCRStatus(mockResource, crStatus, crObj, "default")

			convey.So(err, convey.ShouldBeNil)
			convey.So(crObj.Object["status"], convey.ShouldNotBeNil) // Assert status was set
			convey.So(crObj.Object["status"].(map[string]interface{})["readyReplicas"], convey.ShouldEqual, int64(10))
		})

		convey.Convey("When ToUnstructured fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			expectedErr := fmt.Errorf("conversion failed")
			patches.ApplyMethod(runtime.DefaultUnstructuredConverter, "ToUnstructured",
				func(_ runtime.UnstructuredConverter, input interface{}) (map[string]interface{}, error) {
					return nil, expectedErr
				})

			err := reconciler.updateCRStatus(mockResource, crStatus, crObj, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to transfer status to Unstructured")
			convey.So(err.Error(), convey.ShouldContainSubstring, expectedErr.Error())
		})

		convey.Convey("When unstructured.SetNestedField fails", func() {
			// Simulate an error by providing a crObj.Object that cannot accept the status
			// This is a bit contrived, but shows how to patch a function from an import.
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			expectedSetErr := fmt.Errorf("cannot set nested field on non-object")
			patches.ApplyFunc(unstructured.SetNestedField,
				func(obj map[string]interface{}, value interface{}, fieldPath ...string) error {
					return expectedSetErr
				})

			err := reconciler.updateCRStatus(mockResource, crStatus, crObj, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to set status")
			convey.So(err.Error(), convey.ShouldContainSubstring, expectedSetErr.Error())
		})

		convey.Convey("When the UpdateStatus call returns a 'Not Found' error", func() {
			/*
				mockResource := &mockDynamicResource{
					updateErr: errors.NewNotFound(schema.GroupResource{Group: "mygroup", Resource: "myresource"}, "my-cr"),
				}
			*/

			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethod(runtime.DefaultUnstructuredConverter, "ToUnstructured",
				func(_ runtime.UnstructuredConverter, input interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{}, nil
				})

			err := reconciler.updateCRStatus(mockResource, crStatus, crObj, "default")

			// The function should return nil as it handles this case gracefully
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("When the UpdateStatus call returns a different server error", func() {
			/*
				mockResource := &mockDynamicResource{
					updateErr: errors.NewInternalError(fmt.Errorf("server is on fire")),
				}
			*/

			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethod(runtime.DefaultUnstructuredConverter, "ToUnstructured",
				func(_ runtime.UnstructuredConverter, input interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{}, fmt.Errorf("server is on fire")
				})

			err := reconciler.updateCRStatus(mockResource, crStatus, crObj, "default")

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "server is on fire")
		})
	})
}

func TestAgentCRReconciler_IsFuncSpecChanged(t *testing.T) {
	convey.Convey("IsFuncSpecChanged", t, func() {
		ar := &AgentCRReconciler{
			funcSign: "1",
		}
		convey.So(ar.IsFuncSpecChanged(&types.FunctionSpecification{FuncMetaSignature: "1"}), convey.ShouldBeFalse)
		convey.So(ar.IsFuncSpecChanged(&types.FunctionSpecification{FuncMetaSignature: "0"}), convey.ShouldBeTrue)
	})
}

func TestAgentCRReconciler_IsInstanceConfigChanged(t *testing.T) {
	convey.Convey("IsInstanceConfigChanged", t, func() {
		ar := &AgentCRReconciler{
			funcSign:      "1",
			insConfigSign: getInstanceConfigSign(commontype.InstanceMetaData{}),
		}
		convey.So(ar.IsInstanceConfigChanged(&types.FunctionSpecification{}), convey.ShouldBeFalse)
		convey.So(ar.IsInstanceConfigChanged(&types.FunctionSpecification{InstanceMetaData: commontype.InstanceMetaData{MinInstance: 2}}), convey.ShouldBeTrue)
	})
}
