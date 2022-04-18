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
	summaryOpts prometheus.SummaryOpts
	summaryVec  *prometheus.SummaryVec
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
	s.summaryVec = prometheus.NewSummaryVec(s.summaryOpts, s.labelName)
}

func GetSummaryCollector(name, help string, objectives map[float64]float64, labelName []string) (*SummaryMetric, error) {
	summaryMetric := &SummaryMetric{}
	summaryMetricInterface, ok := summaryMetricNameMap.Load(name)
	//1. 先查看之前有没有注册过同名的metric
	if !ok {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		summaryMetric.setAttributes(name, help, objectives, labelName)
		registerErr := Registry.Register(summaryMetric.summaryVec)
		if registerErr != nil {
			return nil, registerErr
		}
		//3. 把拿到的summaryMetric再添加到summaryMetricNameMap中，代表该summaryMetric已经在注册表中注册过了
		summaryMetricNameMap.Store(name, summaryMetric)
	} else {
		summaryMetric, ok = summaryMetricInterface.(*SummaryMetric)
		if !ok {
			return nil, fmt.Errorf("cannot find metric by name %s", name)
		}
		//4. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(summaryMetric.name, summaryMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
	}
	return summaryMetric, nil
}

func (s *SummaryMetric) DoObserve(labelValue []string, metricValue float64) error {
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(s.labelName, labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//监控非时间指标时，可以手动传进来metricValue，进行observe
	s.summaryVec.With(labels).Observe(metricValue)

	return nil
}

func GetSummaryTimer(name, help string, objective map[float64]float64, labelName, labelValue []string) (*prometheus.Timer, error) {
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	summaryMetric, collectorErr := GetSummaryCollector(name, help, objective, labelName)
	if collectorErr != nil {
		return nil, collectorErr
	}

	timer, buildTimerErr := summaryMetric.BuildTimer(labelValue)
	if buildTimerErr != nil {
		return nil, buildTimerErr
	}

	return timer, nil
}

func (s *SummaryMetric) BuildTimer(labelValue []string) (*prometheus.Timer, error) {
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(s.labelName, labelValue)
	if generateLabelErr != nil {
		return nil, generateLabelErr
	}
	//监控时间指标时，生成timer计时器
	timer := prometheus.NewTimer(s.summaryVec.With(labels))

	return timer, nil
}
