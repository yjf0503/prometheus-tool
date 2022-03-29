package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type SummaryMetric struct {
	name        string
	help        string
	objectives  map[float64]float64
	labelName   []string
	labelValue  []string
	summaryOpts prometheus.SummaryOpts
	summaryVec  *prometheus.SummaryVec
	timer       *prometheus.Timer
}

func (s *SummaryMetric) setAttributes(name, help string, objectives map[float64]float64, labelName []string) {
	s.name = name
	s.help = help
	s.objectives = objectives
	s.labelName = labelName
	s.summaryOpts = prometheus.SummaryOpts{
		Name:       s.name,
		Help:       s.help,
		Objectives: s.objectives,
	}
}

func (s *SummaryMetric) CheckAndRegisterCollector(name, help string, objectives map[float64]float64, labelName []string) *SummaryMetric {
	summaryMetric := summaryMetricNameMap[name]
	if summaryMetric == nil {
		summaryMetric = &SummaryMetric{}
		summaryMetric.setAttributes(name, help, objectives, labelName)
		summaryMetric.summaryVec = prometheus.NewSummaryVec(summaryMetric.summaryOpts, summaryMetric.labelName)
		err := Registry.Register(summaryMetric.summaryVec)
		if err != nil {
			fmt.Print(err.Error())
		}
	} else {
		summaryMetric.setAttributes(name, help, objectives, labelName)
	}
	summaryMetricNameMap[name] = summaryMetric

	return summaryMetric
}

func (s *SummaryMetric) DoObserve(labelValue []string, metricValue float64) error {
	s.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(s.labelName, s.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(s.labelName, s.labelValue)
	s.summaryVec.With(labels).Observe(metricValue)

	return nil
}
