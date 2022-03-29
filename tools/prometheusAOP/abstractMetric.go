package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

// Registry 创建一个自定义的注册表
var Registry = prometheus.NewRegistry()

var histogramMetricNames map[string]*HistogramMetric
var summaryMetricNames map[string]*SummaryMetric
var counterMetricNames map[string]*CounterMetric
var gaugeMetricNames map[string]*GaugeMetric

func init() {
	histogramMetricNames = make(map[string]*HistogramMetric, 0)
	summaryMetricNames = make(map[string]*SummaryMetric, 0)
	counterMetricNames = make(map[string]*CounterMetric, 0)
	gaugeMetricNames = make(map[string]*GaugeMetric, 0)
}

func checkLabelNameAndValue(labelName, labelValue []string) error {
	if len(labelName) != len(labelValue) {
		return fmt.Errorf("labelName is incompatible to labelValue, labelName is %s, while labelValue is %s \n", labelName, labelValue)
	}
	return nil
}

func UnregisterCollectors() {
	for _, v := range counterMetricNames {
		Registry.Unregister(v.counterVec)
	}

	for _, v := range gaugeMetricNames {
		Registry.Unregister(v.gaugeVec)
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
