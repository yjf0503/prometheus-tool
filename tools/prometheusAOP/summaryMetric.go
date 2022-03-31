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
	//1. 先查看之前有没有注册过同名的metric
	if summaryMetric == nil {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		summaryMetric = &SummaryMetric{}
		summaryMetric.setAttributes(name, help, objectives, labelName)
		summaryMetric.summaryVec = prometheus.NewSummaryVec(summaryMetric.summaryOpts, summaryMetric.labelName)
		registerErr := Registry.Register(summaryMetric.summaryVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		//3. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(summaryMetric.name, summaryMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		//3.1 更新下新的help和objectives配置项，如果有更新的话
		summaryMetric.setAttributes(name, help, objectives, labelName)
	}
	//4. 把拿到的summaryMetric再添加到summaryMetricNameMap中，代表该summaryMetric已经在注册表中注册过了
	summaryMetricNameMap[name] = summaryMetric

	return summaryMetric, nil
}

func (s *SummaryMetric) DoObserve(labelValue []string, metricValue float64) error {
	s.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(s.labelName, s.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//监控非时间指标时，可以手动传进来metricValue，进行observe
	s.summaryVec.With(labels).Observe(metricValue)

	return nil
}

func (s *SummaryMetric) BuildTimer(labelValue []string) (*prometheus.Timer, error) {
	s.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(s.labelName, s.labelValue)
	if generateLabelErr != nil {
		return nil, generateLabelErr
	}
	//监控时间指标时，生成timer计时器
	timer := prometheus.NewTimer(s.summaryVec.With(labels))

	return timer, nil
}
