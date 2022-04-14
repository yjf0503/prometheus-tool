package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
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

func GetGaugeCollectorAndSetTimer(name, help string, labelName []string) (*GaugeMetric, time.Time) {
	return GetGaugeCollector(name, help, labelName), time.Now()
}

func GetGaugeCollector(name, help string, labelName []string) *GaugeMetric {
	gaugeMetric := &GaugeMetric{}
	gaugeMetricInterface, ok := gaugeMetricNameMap.Load(name)
	//1. 先查看之前有没有注册过同名的metric
	if !ok {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		gaugeMetric.setAttributes(name, help, labelName, []string{})
		gaugeMetric.gaugeVec = prometheus.NewGaugeVec(gaugeMetric.gaugeOpts, gaugeMetric.labelName)
		registerErr := Registry.Register(gaugeMetric.gaugeVec)
		if registerErr != nil {
			fmt.Println(registerErr.Error())
			//如果有error，返回一个空metric对象
			return &GaugeMetric{}
		}
		//3. 把拿到的gaugeMetric再添加到gaugeMetricNameMap中，代表该gaugeMetric已经在注册表中注册过了
		gaugeMetricNameMap.Store(name, gaugeMetric)
	} else {
		gaugeMetric, ok = gaugeMetricInterface.(*GaugeMetric)
		if !ok {
			err := fmt.Errorf("cannot find metric by name %s", name)
			fmt.Println(err.Error())
			//如果有error，返回一个空metric对象
			return &GaugeMetric{}
		}
		//4. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(gaugeMetric.name, gaugeMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			fmt.Println(checkLabelNamesErr.Error())
			//如果有error，返回一个空metric对象
			return &GaugeMetric{}
		}
	}
	return gaugeMetric
}

func (g *GaugeMetric) DoObserveTimer(labelValue []string, timerStart time.Time) {
	//如果name字段为空，代表metric对象生成有问题，直接返回
	if g.name == "" {
		return
	}
	//计算与timeStart的相隔时间长度，并进行observe
	timerEnd := time.Now()
	timerCost := timerEnd.Sub(timerStart).Microseconds()
	err := g.DoObserve(labelValue, float64(timerCost))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (g *GaugeMetric) DoObserve(labelValue []string, metricValue float64) error {
	g.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(g.labelName, g.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//进行计数操作
	g.gaugeVec.With(labels).Set(metricValue)

	return nil
}
