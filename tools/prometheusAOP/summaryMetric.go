package prometheusAOP

import (
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

func (s *SummaryMetric) GetCollector(name, help string, objectives map[float64]float64, labelName []string) (*SummaryMetric, error) {
	summaryMetric := summaryMetricNameMap[name]
	if summaryMetric == nil {
		summaryMetric = &SummaryMetric{}
		summaryMetric.setAttributes(name, help, objectives, labelName)
		summaryMetric.summaryVec = prometheus.NewSummaryVec(summaryMetric.summaryOpts, summaryMetric.labelName)
		registerErr := Registry.Register(summaryMetric.summaryVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		checkLabelNamesErr := checkLabelNames(summaryMetric.name, summaryMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		summaryMetric.setAttributes(name, help, objectives, labelName)
	}
	summaryMetricNameMap[name] = summaryMetric

	return summaryMetric, nil
}

func (s *SummaryMetric) DoObserve(labelValue []string, metricValue float64) error {
	s.labelValue = labelValue
	checkLabelNameAndValueErr := checkLabelNameAndValue(s.labelName, s.labelValue)
	if checkLabelNameAndValueErr != nil {
		return checkLabelNameAndValueErr
	}

	labels := generateLabels(s.labelName, s.labelValue)

	//1. 监控非时间指标时，可以手动传进来metricValue，进行observe
	s.summaryVec.With(labels).Observe(metricValue)

	////2. 监控时间指标时，可以生成timer计时器，进行observe，将其放进summary指标中去
	//timer := prometheus.NewTimer(s.summaryVec.With(labels))
	////模拟程序执行时间，生成0-999的随机数
	//rand.Seed(time.Now().UnixNano())
	//time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	//timer.ObserveDuration()

	return nil
}
