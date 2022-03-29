package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

// Registry 创建一个自定义的注册表
var Registry = prometheus.NewRegistry()

var histogramMetricNames map[string]bool
var summaryMetricNames map[string]bool
var CounterMetricNames map[string]*CounterMetric
var gaugeMetricNames map[string]bool

func init() {
	histogramMetricNames = make(map[string]bool, 0)
	summaryMetricNames = make(map[string]bool, 0)
	CounterMetricNames = make(map[string]*CounterMetric, 0)
	gaugeMetricNames = make(map[string]bool, 0)
}

func checkLabelNameAndValue(labelName, labelValue []string) error {
	if len(labelName) != len(labelValue) {
		return fmt.Errorf("labelName is incompatible to labelValue, labelName is %s, while labelValue is %s \n", labelName, labelValue)
	}
	return nil
}

func UnregisterCollectors() {
	for _, v := range CounterMetricNames {
		Registry.Unregister(v.CounterVec)
	}
}

func generateLabels(labelName, labelValue []string) map[string]string {
	labels := map[string]string{}
	for k, v := range labelName {
		labels[v] = labelValue[k]
	}
	return labels
}

type metricObject interface {
	buildMetric()                      //创建指标对象
	DoObserve([]string, float64) error //给指标对象注入label值和指标值，对指标进行监控，如果label名称和label值不是一一对应，返回error
}
