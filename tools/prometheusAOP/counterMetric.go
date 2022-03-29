package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type CounterMetric struct {
	Name        string
	Help        string
	LabelName   []string
	labelValue  []string
	CounterOpts prometheus.CounterOpts
	CounterVec  *prometheus.CounterVec
}

func (c *CounterMetric) SetAttributes(name string, help string, labelName []string) {
	c.Name = name
	c.Help = help
	c.LabelName = labelName
	c.CounterOpts = prometheus.CounterOpts{
		Name: c.Name,
		Help: c.Help,
	}

}

func (c *CounterMetric) CheckAndRegisterCollector(name string, help string, labelName string) *CounterMetric {
	counterMetric := CounterMetricNames[name]
	if counterMetric == nil {
		counterMetric = &CounterMetric{}
		counterMetric.SetAttributes(name, help, []string{labelName})
		counterMetric.CounterVec = prometheus.NewCounterVec(counterMetric.CounterOpts, counterMetric.LabelName)
		err := Registry.Register(counterMetric.CounterVec)
		if err != nil {
			fmt.Print(err.Error())
		}
		CounterMetricNames[name] = counterMetric
	} else {
		counterMetric.SetAttributes(name, help, []string{labelName})
	}

	return counterMetric
}

func (c *CounterMetric) DoObserve(labelValue []string, metricValue float64) error {
	c.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(c.LabelName, c.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	if metricValue < 0 {
		metricValue = 0
	}
	labels := generateLabels(c.LabelName, c.labelValue)
	c.CounterVec.With(labels).Add(metricValue)

	return nil
}
