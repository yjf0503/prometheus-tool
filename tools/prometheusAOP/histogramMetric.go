package prometheusAOP

import (
	"github.com/prometheus/client_golang/prometheus"
)

type HistogramMetric struct {
	name         string
	help         string
	buckets      []float64
	labelName    []string
	labelValue   []string
	HistogramVec *prometheus.HistogramVec
	observer     prometheus.Observer
	timer        *prometheus.Timer
}

func (h *HistogramMetric) buildMetric() {
	histogramOpts := prometheus.HistogramOpts{
		Name:    h.name,
		Help:    h.help,
		Buckets: h.buckets,
	}
	h.HistogramVec = prometheus.NewHistogramVec(histogramOpts, h.labelName)
}

func (h *HistogramMetric) DoObserve(labelValue []string, metricValue float64) error {
	h.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(h.labelName, h.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(h.labelName, h.labelValue)
	h.observer = h.HistogramVec.With(labels)
	h.observer.Observe(metricValue)

	return nil
}

func (h *HistogramMetric) Before(name string, help string, buckets []float64, labelName []string) {
	h.name = name
	h.help = help
	h.buckets = buckets
	h.labelName = labelName
	h.buildMetric()

	_ = Registry.Register(h.HistogramVec)
}
