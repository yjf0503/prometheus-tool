package prometheusAOP

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeMetric struct {
	name       string
	help       string
	labelName  []string
	labelValue []string
	labels     prometheus.Labels
	gaugeOpts  prometheus.GaugeOpts
	gaugeVec   *prometheus.GaugeVec
	gauge      prometheus.Gauge
}

func (g *GaugeMetric) setAttributes(name, help string, labelName, labelValue []string) {
	g.name = name
	g.help = help
	g.labelName = labelName
	g.labelValue = labelValue
	gaugeOpts := prometheus.GaugeOpts{
		Name: g.name,
		Help: g.help,
	}

	//labelValue长度大于0，代表要生成的不是gaugeVec，而是有timer的gauge，所以要预先配置好ConstLabels
	if len(labelValue) > 0 {
		//生成后续监控要用到的labelName和labelValue的映射
		labels, generateLabelErr := generateLabels(g.labelName, g.labelValue)
		if generateLabelErr != nil {
			return
		}
		g.labels = labels
		gaugeOpts.ConstLabels = g.labels
	}

	g.gaugeOpts = gaugeOpts
}

func (g *GaugeMetric) GetGaugeVecCollector(name, help string, labelName []string) (*GaugeMetric, error) {
	gaugeMetric := &GaugeMetric{}
	gaugeMetricInterface, ok := gaugeMetricNameMap.Load(name)
	//1. 先查看之前有没有注册过同名的metric
	if !ok {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		gaugeMetric.setAttributes(name, help, labelName, []string{})
		gaugeMetric.gaugeVec = prometheus.NewGaugeVec(gaugeMetric.gaugeOpts, gaugeMetric.labelName)
		registerErr := Registry.Register(gaugeMetric.gaugeVec)
		if registerErr != nil {
			return nil, registerErr
		}
		//3. 把拿到的gaugeMetric再添加到gaugeMetricNameMap中，代表该gaugeMetric已经在注册表中注册过了
		gaugeMetricNameMap.Store(name, gaugeMetric)
	} else {
		gaugeMetric = gaugeMetricInterface.(*GaugeMetric)
		//4. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(gaugeMetric.name, gaugeMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
	}
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

func GetGaugeCollector(name, help string, labelName, labelValue []string) (*GaugeMetric, error) {
	//生成一个新的metric，再注册到Registry中
	gaugeMetric := &GaugeMetric{}
	gaugeMetric.setAttributes(name, help, labelName, labelValue)
	gauge := prometheus.NewGauge(gaugeMetric.gaugeOpts)
	gaugeMetric.gauge = gauge
	registerErr := Registry.Register(gauge)
	if registerErr != nil {
		return nil, registerErr
	}

	return gaugeMetric, nil
}

func (g *GaugeMetric) BuildTimer() *prometheus.Timer {
	//监控时间指标时，生成timer计时器
	gauge := g.gauge
	return prometheus.NewTimer(prometheus.ObserverFunc(gauge.Set))
}

func GetGaugeTimer(name, help string, labelName, labelValue []string) (*prometheus.Timer, error) {
	//获取gauge collector
	gaugeMetric, collectorErr := GetGaugeCollector(name, help, labelName, labelValue)
	if collectorErr != nil {
		return nil, collectorErr
	}

	return gaugeMetric.BuildTimer(), nil
}
