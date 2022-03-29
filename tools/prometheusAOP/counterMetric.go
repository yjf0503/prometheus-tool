package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type CounterMetric struct {
	name        string
	help        string
	labelName   []string
	labelValue  []string
	counterOpts prometheus.CounterOpts
	counterVec  *prometheus.CounterVec
}

func (c *CounterMetric) setAttributes(name string, help string, labelName []string) {
	c.name = name
	c.help = help
	c.labelName = labelName
	c.counterOpts = prometheus.CounterOpts{
		Name: c.name,
		Help: c.help,
	}
}

func (c *CounterMetric) CheckAndRegisterCollector(name string, help string, labelName []string) *CounterMetric {
	counterMetric := counterMetricNames[name]
	if counterMetric == nil {
		counterMetric = &CounterMetric{}
		counterMetric.setAttributes(name, help, labelName)
		counterMetric.counterVec = prometheus.NewCounterVec(counterMetric.counterOpts, counterMetric.labelName)
		err := Registry.Register(counterMetric.counterVec)
		if err != nil {
			fmt.Print(err.Error())
		}
		counterMetricNames[name] = counterMetric
	} else {
		counterMetric.setAttributes(name, help, labelName)
	}

	return counterMetric
}

func (c *CounterMetric) DoObserve(labelValue []string, metricValue float64) error {
	c.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(c.labelName, c.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	if metricValue < 0 {
		metricValue = 0
	}
	labels := generateLabels(c.labelName, c.labelValue)
	c.counterVec.With(labels).Add(metricValue)

	return nil
}
