package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

// Registry 创建一个自定义的注册表
var Registry = prometheus.NewRegistry()

var histogramMetricNameMap map[string]*HistogramMetric
var summaryMetricNameMap map[string]*SummaryMetric
var counterMetricNameMap map[string]*CounterMetric
var gaugeMetricNameMap map[string]*GaugeMetric

func init() {
	histogramMetricNameMap = make(map[string]*HistogramMetric, 0)
	summaryMetricNameMap = make(map[string]*SummaryMetric, 0)
	counterMetricNameMap = make(map[string]*CounterMetric, 0)
	gaugeMetricNameMap = make(map[string]*GaugeMetric, 0)
}

func checkLabelNames(name string, originalLabelName, inputLabelValue []string) error {
	originalLabelNameString := strings.Join(originalLabelName, ",")
	inputLabelNameString := strings.Join(inputLabelValue, ",")

	if originalLabelNameString != inputLabelNameString {
		return fmt.Errorf("labelNames are not same, original labelName of metric {%s} is [%s], while input labelName is [%s] \n", name, originalLabelNameString, inputLabelNameString)
	}

	return nil
}

func checkLabelNameAndValue(labelName, labelValue []string) error {
	if len(labelName) != len(labelValue) {
		return fmt.Errorf("labelName is incompatible to labelValue, labelName is %s, while labelValue is %s \n", labelName, labelValue)
	}
	return nil
}

func generateLabels(labelName, labelValue []string) map[string]string {
	labels := map[string]string{}
	for k, v := range labelName {
		labels[v] = labelValue[k]
	}
	return labels
}

func UnregisterCollectors() {
	for _, v := range counterMetricNameMap {
		Registry.Unregister(v.counterVec)
	}

	for _, v := range gaugeMetricNameMap {
		Registry.Unregister(v.gaugeVec)
	}

	for _, v := range histogramMetricNameMap {
		Registry.Unregister(v.histogramVec)
	}

	for _, v := range summaryMetricNameMap {
		Registry.Unregister(v.summaryVec)
	}
}

type metricObject interface {
	buildMetric()                      //创建指标对象
	DoObserve([]string, float64) error //给指标对象注入label值和指标值，对指标进行监控，如果label名称和label值不是一一对应，返回error
}
