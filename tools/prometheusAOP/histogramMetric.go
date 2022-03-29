package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type HistogramMetric struct {
	name          string
	help          string
	buckets       []float64
	labelName     []string
	labelValue    []string
	histogramOpts prometheus.HistogramOpts
	histogramVec  *prometheus.HistogramVec
	timer         *prometheus.Timer
}

func (h *HistogramMetric) setAttributes(name, help string, buckets []float64, labelName []string) {
	h.name = name
	h.help = help
	h.buckets = buckets
	h.labelName = labelName
	h.histogramOpts = prometheus.HistogramOpts{
		Name:    h.name,
		Help:    h.help,
		Buckets: h.buckets,
	}
}

func (h *HistogramMetric) CheckAndRegisterCollector(name, help string, buckets []float64, labelName []string) *HistogramMetric {
	histogramMetric := histogramMetricNameMap[name]
	if histogramMetric == nil {
		histogramMetric = &HistogramMetric{}
		histogramMetric.setAttributes(name, help, buckets, labelName)
		histogramMetric.histogramVec = prometheus.NewHistogramVec(histogramMetric.histogramOpts, histogramMetric.labelName)
		err := Registry.Register(histogramMetric.histogramVec)
		if err != nil {
			fmt.Print(err.Error())
		}
	} else {
		histogramMetric.setAttributes(name, help, buckets, labelName)
	}
	histogramMetricNameMap[name] = histogramMetric

	return histogramMetric
}

func (h *HistogramMetric) DoObserve(labelValue []string, metricValue float64) error {
	h.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(h.labelName, h.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(h.labelName, h.labelValue)
	h.histogramVec.With(labels).Observe(metricValue)

	return nil
}
