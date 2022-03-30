package prometheusAOP

import (
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

func (c *CounterMetric) setAttributes(name, help string, labelName []string) {
	c.name = name
	c.help = help
	c.labelName = labelName
	c.counterOpts = prometheus.CounterOpts{
		Name: c.name,
		Help: c.help,
	}
}

func (c *CounterMetric) GetCollector(name, help string, labelName []string) (*CounterMetric, error) {
	counterMetric := counterMetricNameMap[name]
	if counterMetric == nil {
		counterMetric = &CounterMetric{}
		counterMetric.setAttributes(name, help, labelName)
		counterMetric.counterVec = prometheus.NewCounterVec(counterMetric.counterOpts, counterMetric.labelName)
		registerErr := Registry.Register(counterMetric.counterVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		checkLabelNamesErr := checkLabelNames(counterMetric.name, counterMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		counterMetric.setAttributes(name, help, labelName)
	}
	counterMetricNameMap[name] = counterMetric

	return counterMetric, nil
}

func (c *CounterMetric) DoObserve(labelValue []string, metricValue float64) error {
	c.labelValue = labelValue
	checkLabelNameAndValueErr := checkLabelNameAndValue(c.labelName, c.labelValue)
	if checkLabelNameAndValueErr != nil {
		return checkLabelNameAndValueErr
	}

	if metricValue < 0 {
		metricValue = 0
	}
	labels := generateLabels(c.labelName, c.labelValue)
	c.counterVec.With(labels).Add(metricValue)

	return nil
}
