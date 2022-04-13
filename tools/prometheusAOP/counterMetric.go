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

func GetCounterCollector(name, help string, labelName []string) (*CounterMetric, error) {
	counterMetric := &CounterMetric{}
	counterMetricInterface, ok := counterMetricNameMap.Load(name)
	//1. 先查看之前有没有注册过同名的metric
	if !ok {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		counterMetric.setAttributes(name, help, labelName)
		counterMetric.counterVec = prometheus.NewCounterVec(counterMetric.counterOpts, counterMetric.labelName)
		registerErr := Registry.Register(counterMetric.counterVec)
		if registerErr != nil {
			return nil, registerErr
		}
		//3. 把拿到的counterMetric再添加到counterMetricNameMap中，代表该counterMetric已经在注册表中注册过了
		counterMetricNameMap.Store(name, counterMetric)
	} else {
		counterMetric = counterMetricInterface.(*CounterMetric)
		//4. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(counterMetric.name, counterMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
	}
	return counterMetric, nil
}

func (c *CounterMetric) DoObserve(labelValue []string, metricValue float64) error {
	c.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(c.labelName, c.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}

	if metricValue < 0 {
		metricValue = 0
	}
	//进行计数操作
	c.counterVec.With(labels).Add(metricValue)

	return nil
}
