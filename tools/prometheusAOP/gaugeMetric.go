package prometheusAOP

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeMetric struct {
	name       string
	help       string
	labelName  []string
	labelValue []string
	gaugeOpts  prometheus.GaugeOpts
	gaugeVec   *prometheus.GaugeVec
}

func (g *GaugeMetric) setAttributes(name, help string, labelName []string) {
	g.name = name
	g.help = help
	g.labelName = labelName
	g.gaugeOpts = prometheus.GaugeOpts{
		Name: g.name,
		Help: g.help,
	}
}

func (g *GaugeMetric) GetCollector(name, help string, labelName []string) (*GaugeMetric, error) {
	gaugeMetric := gaugeMetricNameMap[name]
	if gaugeMetric == nil {
		gaugeMetric = &GaugeMetric{}
		gaugeMetric.setAttributes(name, help, labelName)
		gaugeMetric.gaugeVec = prometheus.NewGaugeVec(gaugeMetric.gaugeOpts, gaugeMetric.labelName)
		registerErr := Registry.Register(gaugeMetric.gaugeVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		checkLabelNamesErr := checkLabelNames(gaugeMetric.name, gaugeMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		gaugeMetric.setAttributes(name, help, labelName)
	}
	gaugeMetricNameMap[name] = gaugeMetric

	return gaugeMetric, nil
}

func (g *GaugeMetric) DoObserve(labelValue []string, metricValue float64) error {
	g.labelValue = labelValue
	labels, generateLabelErr := generateLabels(g.labelName, g.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	g.gaugeVec.With(labels).Add(metricValue)

	return nil
}
