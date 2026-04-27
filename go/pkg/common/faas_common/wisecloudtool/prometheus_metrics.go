/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

// Package wisecloudtool -
package wisecloudtool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	k8stype "k8s.io/apimachinery/pkg/types"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

const defaultLabel = "UNKNOWN_LABEL"
const labelLen = 8

var (
	concurrencyGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuanrong_concurrency_num",
			Help: "The current concurrency number of the application.",
		},
		[]string{"businessid", "tenantid", "funcname", "version", "label", "namespace", "deployment_name", "pod_name"},
	)

	leaseRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuanrong_lease_total",
			Help: "The lease total number of the application.",
		},
		[]string{"businessid", "tenantid", "funcname", "version", "label", "namespace", "deployment_name", "pod_name"},
	)
)

// GetLeaseRequestTotal -
func GetLeaseRequestTotal() *prometheus.CounterVec {
	return leaseRequestTotal
}

// GetConcurrencyGauge -
func GetConcurrencyGauge() *prometheus.GaugeVec {
	return concurrencyGauge
}

// MetricProvider -
type MetricProvider struct {
	sync.RWMutex
	// key is {funcKey}#{invokeLabel}, subKey namespace value is {namespace, podName}
	WorkLoadMap map[string]map[string]*k8stype.NamespacedName
}

// NewMetricProvider -
func NewMetricProvider() *MetricProvider {
	return &MetricProvider{
		RWMutex:     sync.RWMutex{},
		WorkLoadMap: make(map[string]map[string]*k8stype.NamespacedName),
	}
}

// AddWorkLoad -
func (m *MetricProvider) AddWorkLoad(funcKey string, invokeLabel string, namespaceName *k8stype.NamespacedName) {
	workload := getWorkloadName(funcKey, invokeLabel)
	m.Lock()
	defer m.Unlock()

	deployments, ok := m.WorkLoadMap[workload]
	if !ok {
		deployments = make(map[string]*k8stype.NamespacedName)
		m.WorkLoadMap[workload] = deployments
	}
	if _, ok = deployments[namespaceName.String()]; !ok {
		deployments[namespaceName.String()] = namespaceName
	}
}

// EnsureConcurrencyGaugeWithLabel -
func (m *MetricProvider) EnsureConcurrencyGaugeWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}

	m.RLock()
	defer m.RUnlock()
	_, err := concurrencyGauge.GetMetricWithLabelValues(labels...)
	return err
}

// EnsureLeaseRequestTotalWithLabel -
func (m *MetricProvider) EnsureLeaseRequestTotalWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}

	m.RLock()
	defer m.RUnlock()
	_, err := leaseRequestTotal.GetMetricWithLabelValues(labels...)
	return err
}

// Exist -
func (m *MetricProvider) Exist(funcKey string, invokeLabel string) bool {
	return m.GetRandomDeployment(funcKey, invokeLabel) != nil
}

// GetRandomDeployment -
func (m *MetricProvider) GetRandomDeployment(funcKey string, invokeLabel string) *k8stype.NamespacedName {
	workName := getWorkloadName(funcKey, invokeLabel)
	m.RLock()
	defer m.RUnlock()
	deployments, ok := m.WorkLoadMap[workName]
	if !ok {
		return nil
	}
	if len(deployments) == 0 {
		return nil
	}
	for _, namespaceName := range deployments {
		return namespaceName
	}
	return nil
}

// IncLeaseRequestTotalWithLabel -
func (m *MetricProvider) IncLeaseRequestTotalWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}
	counter, err := leaseRequestTotal.GetMetricWithLabelValues(labels...)
	if err != nil {
		return err
	}
	counter.Inc()
	return nil
}

// IncConcurrencyGaugeWithLabel -
func (m *MetricProvider) IncConcurrencyGaugeWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}
	gauge, err := concurrencyGauge.GetMetricWithLabelValues(labels...)
	if err != nil {
		return err
	}
	gauge.Inc()
	return nil
}

// DecConcurrencyGaugeWithLabel -
func (m *MetricProvider) DecConcurrencyGaugeWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}
	gauge, err := concurrencyGauge.GetMetricWithLabelValues(labels...)
	if err != nil {
		return err
	}
	gauge.Dec()
	return nil
}

// ClearConcurrencyGaugeWithLabel -
func (m *MetricProvider) ClearConcurrencyGaugeWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}
	concurrencyGauge.DeleteLabelValues(labels...)
	return nil
}

// ClearLeaseRequestTotalWithLabel -
func (m *MetricProvider) ClearLeaseRequestTotalWithLabel(labels []string) error {
	if len(labels) != labelLen {
		return fmt.Errorf("labels len must be 8")
	}
	leaseRequestTotal.DeleteLabelValues(labels...)
	return nil
}

// ClearMetricsForFunction -
func (m *MetricProvider) ClearMetricsForFunction(funcMetaData *types.FuncMetaData) {
	funcKey0 := urnutils.CombineFunctionKey(funcMetaData.TenantID, funcMetaData.FuncName, funcMetaData.Version)
	m.Lock()
	defer m.Unlock()
	for workload, _ := range m.WorkLoadMap {
		funcKey1, invokeLabel := GetFuncKeyAndLabelFromWorkload(workload)
		if funcKey0 == funcKey1 {
			m.clearMetricsForInsConfigWithoutLock(funcMetaData, invokeLabel)
		}
	}
}

// ClearMetricsForInsConfig -
func (m *MetricProvider) ClearMetricsForInsConfig(funcMetaData *types.FuncMetaData, invokeLabel string) {
	m.Lock()
	m.clearMetricsForInsConfigWithoutLock(funcMetaData, invokeLabel)
	m.Unlock()
}

func (m *MetricProvider) clearMetricsForInsConfigWithoutLock(funcMetaData *types.FuncMetaData, invokeLabel string) {
	// 得看下和FunctionVersion有啥区别
	funcKey := urnutils.CombineFunctionKey(funcMetaData.TenantID, funcMetaData.FuncName, funcMetaData.Version)
	workload := getWorkloadName(funcKey, invokeLabel)
	deployments, ok := m.WorkLoadMap[workload]
	if !ok {
		return
	}
	delete(m.WorkLoadMap, workload)

	if invokeLabel == "" {
		invokeLabel = defaultLabel
	}

	for _, deployment := range deployments {
		labels := map[string]string{
			"businessid":      funcMetaData.BusinessID,
			"tenantid":        funcMetaData.TenantID,
			"funcname":        funcMetaData.FuncName,
			"version":         funcMetaData.Version,
			"label":           invokeLabel,
			"namespace":       deployment.Namespace,
			"deployment_name": deployment.Name,
		}
		concurrencyGauge.DeletePartialMatch(labels)
		leaseRequestTotal.DeletePartialMatch(labels)
	}
}

// GetMetricLabels -
// 判断label是否符合预期
func GetMetricLabels(funcMetaData *types.FuncMetaData, invokeLabel string,
	namespace string, deploymentName string, podName string) []string {
	var metricLabelValue []string
	if namespace != "" && deploymentName != "" && podName != "" && funcMetaData != nil {
		if invokeLabel == "" {
			invokeLabel = defaultLabel
		}
		metricLabelValue = []string{
			funcMetaData.BusinessID,
			funcMetaData.TenantID,
			funcMetaData.FuncName,
			funcMetaData.Version,
			invokeLabel,
			namespace,
			deploymentName,
			podName}
	}
	return metricLabelValue
}

// GetFuncKeyAndLabelFromWorkload -
func GetFuncKeyAndLabelFromWorkload(workload string) (string, string) {
	strs := strings.Split(workload, "#")
	if len(strs) == 2 { // deployment key must be 2
		return strs[0], strs[1]
	}
	return "", ""
}

// getWorkloadName - get deploymentforfunckey
func getWorkloadName(funcKey, invokeLabel string) string {
	if invokeLabel == "" {
		invokeLabel = defaultLabel
	}
	return fmt.Sprintf("%s#%s", funcKey, invokeLabel)
}
