package prometheusAOP

import (
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

func (g *GaugeMetric) setAttributes(name, help string, labelName []string) {
	g.name = name
	g.help = help
	g.labelName = labelName
	g.gaugeOpts = prometheus.GaugeOpts{
		Name: g.name,
		Help: g.help,
	}
}

func (g *GaugeMetric) GetCollector(name, help string, labelName []string) (*GaugeMetric, error) {
	gaugeMetric := gaugeMetricNameMap[name]
	//1. 先查看之前有没有注册过同名的metric
	if gaugeMetric == nil {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		gaugeMetric = &GaugeMetric{}
		gaugeMetric.setAttributes(name, help, labelName)
		gaugeMetric.gaugeVec = prometheus.NewGaugeVec(gaugeMetric.gaugeOpts, gaugeMetric.labelName)
		registerErr := Registry.Register(gaugeMetric.gaugeVec)
		if registerErr != nil {
			return nil, registerErr
		}
	} else {
		//3. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(gaugeMetric.name, gaugeMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
		//3.1 更新下新的help配置项，如果有更新的话
		gaugeMetric.setAttributes(name, help, labelName)
	}
	//4. 把拿到的gaugeMetric再添加到gaugeMetricNameMap中，代表该gaugeMetric已经在注册表中注册过了
	gaugeMetricNameMap[name] = gaugeMetric

	return gaugeMetric, nil
}

func (g *GaugeMetric) DoObserve(labelValue []string, metricValue float64) error {
	g.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(g.labelName, g.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//进行计数操作
	g.gaugeVec.With(labels).Add(metricValue)

	return nil
}
