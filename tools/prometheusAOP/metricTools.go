package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
)

// Registry 创建一个自定义的注册表
//var Registry = prometheus.NewRegistry()
var Registry = prometheus.DefaultRegisterer

// 记录本进程生命周期内创建的各类指标，避免重新注册
var histogramMetricNameMap sync.Map
var summaryMetricNameMap sync.Map
var counterMetricNameMap sync.Map
var gaugeMetricNameMap sync.Map

func init() {
	histogramMetricNameMap = sync.Map{}
	summaryMetricNameMap = sync.Map{}
	counterMetricNameMap = sync.Map{}
	gaugeMetricNameMap = sync.Map{}
}

//检测指标已注册的labelName和传入的labelName是否相同，不同的话返回error
func checkLabelNames(name string, originalLabelName, inputLabelValue []string) error {
	originalLabelNameString := strings.Join(originalLabelName, ",")
	inputLabelNameString := strings.Join(inputLabelValue, ",")

	if originalLabelNameString != inputLabelNameString {
		return fmt.Errorf("labelNames are not same, original labelName of metric {%s} is [%s], while input labelName is [%s] \n", name, originalLabelNameString, inputLabelNameString)
	}

	return nil
}

//检测指标的labelName和labelValue是否匹配，不匹配的话返回error
func checkLabelNameAndValue(labelName, labelValue []string) error {
	if len(labelName) != len(labelValue) {
		return fmt.Errorf("labelName is incompatible to labelValue, labelName is %s, while labelValue is %s \n", labelName, labelValue)
	}
	return nil
}

//生成labelName和labelValue的映射
func generateLabels(labelName, labelValue []string) (map[string]string, error) {
	checkLabelNameAndValueErr := checkLabelNameAndValue(labelName, labelValue)
	if checkLabelNameAndValueErr != nil {
		return nil, checkLabelNameAndValueErr
	}

	labels := map[string]string{}
	for k, v := range labelName {
		labels[v] = labelValue[k]
	}

	return labels, nil
}

func UnregisterCollectors() {
	counterMetricNameMap.Range(func(k, v interface{}) bool {
		counterMetric, ok := v.(*CounterMetric)
		if !ok {
			return false
		}
		Registry.Unregister(counterMetric.counterVec)
		return true
	})

	gaugeMetricNameMap.Range(func(k, v interface{}) bool {
		gaugeMetric, ok := v.(*GaugeMetric)
		if !ok {
			return false
		}
		Registry.Unregister(gaugeMetric.gaugeVec)
		return true
	})

	histogramMetricNameMap.Range(func(k, v interface{}) bool {
		histogramMetric, ok := v.(*HistogramMetric)
		if !ok {
			return false
		}
		Registry.Unregister(histogramMetric.histogramVec)
		return true
	})

	summaryMetricNameMap.Range(func(k, v interface{}) bool {
		summaryMetric, ok := v.(*SummaryMetric)
		if !ok {
			return false
		}
		Registry.Unregister(summaryMetric.summaryVec)
		return true
	})
}
