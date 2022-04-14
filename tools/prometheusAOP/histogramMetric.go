package prometheusAOP

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type HistogramMetric struct {
	name          string
	help          string
	buckets       []float64
	labelName     []string
	labelValue    []string
	histogramOpts prometheus.HistogramOpts
	histogramVec  *prometheus.HistogramVec
}

func (h *HistogramMetric) setAttributes(name, help string, buckets []float64, labelName []string) {
	h.name = name
	h.help = help
	h.buckets = buckets
	h.labelName = labelName
	h.histogramOpts = prometheus.HistogramOpts{
		Name:    h.name,
		Help:    h.help,
		Buckets: h.buckets,
	}
}

func GetHistogramCollector(name, help string, buckets []float64, labelName []string) (*HistogramMetric, error) {
	histogramMetric := &HistogramMetric{}
	histogramMetricInterface, ok := histogramMetricNameMap.Load(name)
	//1. 先查看之前有没有注册过同名的metric
	if !ok {
		//2. 如果之前没注册过，生成一个新的，再注册到自定义Registry中
		histogramMetric.setAttributes(name, help, buckets, labelName)
		histogramMetric.histogramVec = prometheus.NewHistogramVec(histogramMetric.histogramOpts, histogramMetric.labelName)
		registerErr := Registry.Register(histogramMetric.histogramVec)
		if registerErr != nil {
			return nil, registerErr
		}
		//3. 把拿到的histogramMetric再添加到histogramMetricNameMap中，代表该histogramMetric已经在注册表中注册过了
		histogramMetricNameMap.Store(name, histogramMetric)
	} else {
		histogramMetric, ok = histogramMetricInterface.(*HistogramMetric)
		if !ok {
			return nil, fmt.Errorf("cannot find metric by name %s", name)
		}
		//4. 如果之前注册过同名的metric，需要检测下新传进来的labelName和之前的一不一致，必须保持一致，不然会返回error
		checkLabelNamesErr := checkLabelNames(histogramMetric.name, histogramMetric.labelName, labelName)
		if checkLabelNamesErr != nil {
			return nil, checkLabelNamesErr
		}
	}
	return histogramMetric, nil
}

func (h *HistogramMetric) DoObserve(labelValue []string, metricValue float64) error {
	h.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(h.labelName, h.labelValue)
	if generateLabelErr != nil {
		return generateLabelErr
	}
	//监控非时间指标时，可以手动传进来metricValue，进行observe
	h.histogramVec.With(labels).Observe(metricValue)

	return nil
}

func GetHistogramTimer(name, help string, buckets []float64, labelName, labelValue []string) (*prometheus.Timer, error) {
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	histogramMetric, collectorErr := GetHistogramCollector(name, help, buckets, labelName)
	if collectorErr != nil {
		return nil, collectorErr
	}

	timer, buildTimerErr := histogramMetric.buildTimer(labelValue)
	if buildTimerErr != nil {
		return nil, buildTimerErr
	}

	return timer, nil
}

func (h *HistogramMetric) buildTimer(labelValue []string) (*prometheus.Timer, error) {
	h.labelValue = labelValue
	//生成后续监控要用到的labelName和labelValue的映射
	labels, generateLabelErr := generateLabels(h.labelName, h.labelValue)
	if generateLabelErr != nil {
		return nil, generateLabelErr
	}
	//监控时间指标时，生成timer计时器
	timer := prometheus.NewTimer(h.histogramVec.With(labels))

	return timer, nil
}
