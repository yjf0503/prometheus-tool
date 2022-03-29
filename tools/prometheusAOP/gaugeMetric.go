package prometheusAOP

import (
	"fmt"
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

func (g *GaugeMetric) setAttributes(name string, help string, labelName []string) {
	g.name = name
	g.help = help
	g.labelName = labelName
	g.gaugeOpts = prometheus.GaugeOpts{
		Name: g.name,
		Help: g.help,
	}
}

func (g *GaugeMetric) CheckAndRegisterCollector(name string, help string, labelName []string) *GaugeMetric {
	gaugeMetric := gaugeMetricNames[name]
	if gaugeMetric == nil {
		gaugeMetric = &GaugeMetric{}
		gaugeMetric.setAttributes(name, help, labelName)
		gaugeMetric.gaugeVec = prometheus.NewGaugeVec(gaugeMetric.gaugeOpts, gaugeMetric.labelName)
		err := Registry.Register(gaugeMetric.gaugeVec)
		if err != nil {
			fmt.Print(err.Error())
		}
		gaugeMetricNames[name] = gaugeMetric
	} else {
		gaugeMetric.setAttributes(name, help, labelName)
	}

	return gaugeMetric
}

func (g *GaugeMetric) DoObserve(labelValue []string, metricValue float64) error {
	g.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(g.labelName, g.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(g.labelName, g.labelValue)
	g.gaugeVec.With(labels).Add(metricValue)

	return nil
}
