package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"testing"
	"time"
)

var histogramMetricName = "request_histogram_total"
var histogramMetricHelp = "test request histogram"
var requestTimeBucket = []float64{0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
var requestTime = []float64{0.1, 0.15, 0.2, 0.23, 0.25, 0.4, 0.5, 0.7, 0.85, 0.9}

func TestHistogramMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//收集指标
			err := doHistogramObserve(histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue, requestTime[i])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "secondGoroutine"}
			//收集指标
			err := doHistogramObserve(histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue, requestTime[i])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	select {}
}

func doHistogramObserve(name, help string, buckets []float64, labelName, labelValue []string, metricValue float64) error {
	histogramMetric := &prometheusAOP.HistogramMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	histogramMetric, collectorErr := histogramMetric.GetCollector(name, help, buckets, labelName)
	if collectorErr != nil {
		return collectorErr
	}

	//执行指标数据收集
	observeErr := histogramMetric.DoObserve(labelValue, metricValue)
	if observeErr != nil {
		return observeErr
	}

	return nil
}
