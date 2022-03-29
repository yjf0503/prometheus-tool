package prometheusAOP

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeMetric struct {
	name       string
	help       string
	labelName  []string
	labelValue []string
	GaugeVec   *prometheus.GaugeVec
}

func (g *GaugeMetric) buildMetric() {
	gaugeOpts := prometheus.GaugeOpts{
		Name: g.name,
		Help: g.help,
	}
	g.GaugeVec = prometheus.NewGaugeVec(gaugeOpts, g.labelName)
}

func (g *GaugeMetric) DoObserve(labelValue []string, metricValue float64) error {
	g.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(g.labelName, g.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(g.labelName, g.labelValue)
	g.GaugeVec.With(labels).Add(metricValue)

	return nil
}

func (g *GaugeMetric) Before(name string, help string, labelName []string) {
	g.name = name
	g.help = help
	g.labelName = labelName
	g.buildMetric()

	_ = Registry.Register(g.GaugeVec)
}
