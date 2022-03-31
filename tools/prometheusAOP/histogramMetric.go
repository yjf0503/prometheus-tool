package prometheusAOP

import (
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

func (h *HistogramMetric) GetCollector(name, help string, buckets []float64, labelName []string) (*HistogramMetric, error) {
	histogramMetric := histogramMetricNameMap[name]
	if histogramMetric == nil {
		histogramMetric = &HistogramMetric{}
		histogramMetric.setAttributes(name, help, buckets, labelName)
		histogramMetric.histogramVec = prometheus.NewHistogramVec(histogramMetric.histogramOpts, histogramMetric.labelName)
		registerErr := Registry.Register(histogramMetric.histogramVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		checkLabelNamesErr := checkLabelNames(histogramMetric.name, histogramMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		histogramMetric.setAttributes(name, help, buckets, labelName)
	}
	histogramMetricNameMap[name] = histogramMetric

	return histogramMetric, nil
}

func (h *HistogramMetric) DoObserve(labelValue []string, metricValue float64) error {
	h.labelValue = labelValue
	labels, generateLabelErr := generateLabels(h.labelName, h.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//监控非时间指标时，可以手动传进来metricValue，进行observe
	h.histogramVec.With(labels).Observe(metricValue)

	return nil
}

func (h *HistogramMetric) BuildTimer(labelValue []string) (*prometheus.Timer, error) {
	h.labelValue = labelValue
	labels, generateLabelErr := generateLabels(h.labelName, h.labelValue)
	if generateLabelErr != nil {
		return nil, generateLabelErr
	}
	//监控时间指标时，生成timer计时器
	timer := prometheus.NewTimer(h.histogramVec.With(labels))

	return timer, nil
}
