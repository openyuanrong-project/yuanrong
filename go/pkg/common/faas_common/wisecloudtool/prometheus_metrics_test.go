package wisecloudtool

import (
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/stretchr/testify/assert"
	k8stype "k8s.io/apimachinery/pkg/types"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

func TestNewMetricProvider(t *testing.T) {
	provider := NewMetricProvider()
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.WorkLoadMap)
	assert.Equal(t, 0, len(provider.WorkLoadMap))
}

func TestMetricProvider_AddWorkLoad(t *testing.T) {
	t.Run("Add new workload", func(t *testing.T) {
		provider := NewMetricProvider()
		funcKey := "test-func"
		invokeLabel := "test-label"
		namespaceName := &k8stype.NamespacedName{
			Namespace: "test-ns",
			Name:      "test-name",
		}

		provider.AddWorkLoad(funcKey, invokeLabel, namespaceName)

		assert.Equal(t, 1, len(provider.WorkLoadMap))
		assert.Equal(t, 1, len(provider.WorkLoadMap[getWorkloadName(funcKey, invokeLabel)]))
	})

	t.Run("Add duplicate workload", func(t *testing.T) {
		provider := NewMetricProvider()
		funcKey := "test-func"
		invokeLabel := "test-label"
		namespaceName := &k8stype.NamespacedName{
			Namespace: "test-ns",
			Name:      "test-name",
		}

		// Add twice
		provider.AddWorkLoad(funcKey, invokeLabel, namespaceName)
		provider.AddWorkLoad(funcKey, invokeLabel, namespaceName)

		assert.Equal(t, 1, len(provider.WorkLoadMap))
		assert.Equal(t, 1, len(provider.WorkLoadMap[getWorkloadName(funcKey, invokeLabel)]))
	})
}

func TestMetricProvider_Exist(t *testing.T) {
	provider := NewMetricProvider()
	funcKey := "test-func"
	invokeLabel := "test-label"

	t.Run("Workload does not exist", func(t *testing.T) {
		assert.False(t, provider.Exist(funcKey, invokeLabel))
	})

	t.Run("Workload exists", func(t *testing.T) {
		provider.AddWorkLoad(funcKey, invokeLabel, &k8stype.NamespacedName{
			Namespace: "test-ns",
			Name:      "test-name",
		})
		assert.True(t, provider.Exist(funcKey, invokeLabel))
	})
}

func TestMetricProvider_GetRandomDeployment(t *testing.T) {
	provider := NewMetricProvider()
	funcKey := "test-func"
	invokeLabel := "test-label"
	testDeployment0 := &k8stype.NamespacedName{
		Namespace: "test-ns-0",
		Name:      "test-name-0",
	}

	testDeployment1 := &k8stype.NamespacedName{
		Namespace: "test-ns-1",
		Name:      "test-name-1",
	}

	t.Run("Get non-existent deployment", func(t *testing.T) {
		assert.Nil(t, provider.GetRandomDeployment(funcKey, invokeLabel))
	})

	t.Run("Get existing deployment", func(t *testing.T) {
		provider.AddWorkLoad(funcKey, invokeLabel, testDeployment0)
		provider.AddWorkLoad(funcKey, invokeLabel, testDeployment1)
		flag0 := false
		flag1 := false
		for i := 0; i < 100; i++ {
			result := provider.GetRandomDeployment(funcKey, invokeLabel)
			switch result.Name {
			case "test-name-0":
				flag0 = true
			case "test-name-1":
				flag1 = true
			}
			if flag1 && flag0 {
				break
			}
		}
		assert.True(t, flag0 && flag1)
	})
}

func TestMetricProvider_ClearMetrics(t *testing.T) {
	provider := NewMetricProvider()
	funcMeta := &types.FuncMetaData{
		TenantID:   "tenant1",
		FuncName:   "func1",
		Version:    "v1",
		BusinessID: "biz1",
	}
	invokeLabel := "test-label"
	workload := getWorkloadName(urnutils.CombineFunctionKey(funcMeta.TenantID, funcMeta.FuncName, funcMeta.Version), invokeLabel)

	// Add test data
	provider.AddWorkLoad(
		urnutils.CombineFunctionKey(funcMeta.TenantID, funcMeta.FuncName, funcMeta.Version),
		invokeLabel,
		&k8stype.NamespacedName{
			Namespace: "test-ns",
			Name:      "test-name",
		},
	)

	t.Run("Clear function metrics", func(t *testing.T) {
		provider.ClearMetricsForFunction(funcMeta)
		assert.Equal(t, 0, len(provider.WorkLoadMap))
	})

	t.Run("Clear instance config metrics", func(t *testing.T) {
		// Re-add data
		provider.AddWorkLoad(
			urnutils.CombineFunctionKey(funcMeta.TenantID, funcMeta.FuncName, funcMeta.Version),
			invokeLabel,
			&k8stype.NamespacedName{
				Namespace: "test-ns",
				Name:      "test-name",
			},
		)

		provider.ClearMetricsForInsConfig(funcMeta, invokeLabel)
		assert.Nil(t, provider.WorkLoadMap[workload])
	})
}

func TestGetMetricLabels(t *testing.T) {
	funcMeta := &types.FuncMetaData{
		BusinessID: "biz1",
		TenantID:   "tenant1",
		FuncName:   "func1",
		Version:    "v1",
	}

	t.Run("Generate complete labels", func(t *testing.T) {
		labels := GetMetricLabels(funcMeta, "label1", "ns1", "deploy1", "pod1")
		assert.Equal(t, []string{"biz1", "tenant1", "func1", "v1", "label1", "ns1", "deploy1", "pod1"}, labels)
	})

	t.Run("Use default label", func(t *testing.T) {
		labels := GetMetricLabels(funcMeta, "", "ns1", "deploy1", "pod1")
		assert.Equal(t, "UNKNOWN_LABEL", labels[4])
	})

	t.Run("Return nil when missing required parameters", func(t *testing.T) {
		assert.Nil(t, GetMetricLabels(nil, "label1", "ns1", "deploy1", "pod1"))
		assert.Nil(t, GetMetricLabels(funcMeta, "label1", "", "deploy1", "pod1"))
	})
}

func TestWorkloadHelpers(t *testing.T) {
	t.Run("Get workload name", func(t *testing.T) {
		name := getWorkloadName("func1", "label1")
		assert.Equal(t, "func1#label1", name)
		assert.Equal(t, "func1#UNKNOWN_LABEL", getWorkloadName("func1", ""))
	})

	t.Run("Parse from workload name", func(t *testing.T) {
		funcKey, label := GetFuncKeyAndLabelFromWorkload("func1#label1")
		assert.Equal(t, "func1", funcKey)
		assert.Equal(t, "label1", label)

		funcKey, label = GetFuncKeyAndLabelFromWorkload("invalid")
		assert.Equal(t, "", funcKey)
		assert.Equal(t, "", label)
	})
}

func TestMetricProvider(t *testing.T) {
	convey.Convey("Test MetricProvider Functions", t, func() {
		m := &MetricProvider{}
		validLabels := make([]string, labelLen)
		invalidLabels := make([]string, labelLen-1)

		convey.Convey("Test IncLeaseRequestTotalWithLabel", func() {
			convey.Convey("should return error for invalid label length", func() {
				err := m.IncLeaseRequestTotalWithLabel(invalidLabels)
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "labels len must be 8")
			})

			convey.Convey("should handle GetMetricWithLabelValues error", func() {
				patches := gomonkey.ApplyMethodFunc(leaseRequestTotal, "GetMetricWithLabelValues", func(...string) (prometheus.Counter, error) {
					return nil, fmt.Errorf("mock error")
				})
				defer patches.Reset()

				err := m.IncLeaseRequestTotalWithLabel(validLabels)
				convey.So(err, convey.ShouldNotBeNil)
			})

			convey.Convey("should increment counter successfully", func() {
				patches := gomonkey.ApplyMethodFunc(leaseRequestTotal, "GetMetricWithLabelValues", func(...string) (prometheus.Counter, error) {
					counter := &fakeCounter{}
					return counter, nil
				})
				defer patches.Reset()

				err := m.IncLeaseRequestTotalWithLabel(validLabels)
				convey.So(err, convey.ShouldBeNil)
			})
		})

		convey.Convey("Test IncConcurrencyGaugeWithLabel", func() {
			convey.Convey("should return error for invalid label length", func() {
				err := m.IncConcurrencyGaugeWithLabel(invalidLabels)
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "labels len must be 8")
			})

			convey.Convey("should handle GetMetricWithLabelValues error", func() {
				patches := gomonkey.ApplyMethodFunc(concurrencyGauge, "GetMetricWithLabelValues", func(...string) (prometheus.Gauge, error) {
					return nil, fmt.Errorf("mock error")
				})
				defer patches.Reset()

				err := m.IncConcurrencyGaugeWithLabel(validLabels)
				convey.So(err, convey.ShouldNotBeNil)
			})

			convey.Convey("should increment gauge successfully", func() {
				patches := gomonkey.ApplyMethodFunc(concurrencyGauge, "GetMetricWithLabelValues", func(...string) (prometheus.Gauge, error) {
					gauge := &fakeGauge{}
					return gauge, nil
				})
				defer patches.Reset()

				err := m.IncConcurrencyGaugeWithLabel(validLabels)
				convey.So(err, convey.ShouldBeNil)
			})
		})

		convey.Convey("Test DecConcurrencyGaugeWithLabel", func() {
			convey.Convey("should decrement gauge successfully", func() {
				patches := gomonkey.ApplyMethodFunc(concurrencyGauge, "GetMetricWithLabelValues", func(...string) (prometheus.Gauge, error) {
					gauge := &fakeGauge{}
					return gauge, nil
				})
				defer patches.Reset()

				err := m.DecConcurrencyGaugeWithLabel(validLabels)
				convey.So(err, convey.ShouldBeNil)
			})
		})

		convey.Convey("Test ClearConcurrencyGaugeWithLabel", func() {
			convey.Convey("should clear gauge successfully", func() {
				patches := gomonkey.ApplyMethodFunc(concurrencyGauge, "DeleteLabelValues", func(...string) bool {
					return true
				})
				defer patches.Reset()

				err := m.ClearConcurrencyGaugeWithLabel(validLabels)
				convey.So(err, convey.ShouldBeNil)
			})
		})

		convey.Convey("Test ClearLeaseRequestTotalWithLabel", func() {
			convey.Convey("should clear counter successfully", func() {
				patches := gomonkey.ApplyMethodFunc(leaseRequestTotal, "DeleteLabelValues", func(...string) bool {
					return true
				})
				defer patches.Reset()

				err := m.ClearLeaseRequestTotalWithLabel(validLabels)
				convey.So(err, convey.ShouldBeNil)
			})
		})
	})
}

type fakeCounter struct {
	prometheus.Counter
}

func (f *fakeCounter) Inc() {}

type fakeGauge struct {
	prometheus.Gauge
}

func (f *fakeGauge) Inc() {}
func (f *fakeGauge) Dec() {}
