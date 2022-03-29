package prometheusAOP

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SummaryMetric struct {
	name       string
	help       string
	objectives map[float64]float64
	labelName  []string
	labelValue []string
	SummaryVec *prometheus.SummaryVec
	observer   prometheus.Observer
	timer      *prometheus.Timer
}

func (s *SummaryMetric) buildMetric() {
	summaryOpts := prometheus.SummaryOpts{
		Name:       s.name,
		Help:       s.help,
		Objectives: s.objectives,
	}
	s.SummaryVec = prometheus.NewSummaryVec(summaryOpts, s.labelName)
}

func (s *SummaryMetric) DoObserve(labelValue []string, metricValue float64) error {
	s.labelValue = labelValue
	checkLabelNameAndValueResult := checkLabelNameAndValue(s.labelName, s.labelValue)
	if checkLabelNameAndValueResult != nil {
		return checkLabelNameAndValueResult
	}

	labels := generateLabels(s.labelName, s.labelValue)
	s.observer = s.SummaryVec.With(labels)
	s.observer.Observe(metricValue)

	return nil
}

func (s *SummaryMetric) Before(name string, help string, objectives map[float64]float64, labelName []string) {
	s.name = name
	s.help = help
	s.objectives = objectives
	s.labelName = labelName
	s.buildMetric()

	_ = Registry.Register(s.SummaryVec)
}
