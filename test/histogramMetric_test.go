package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
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
			//收集非时间指标
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
			//收集非时间指标
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

func TestTimerHistogramMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//生成histogram指标的timer
			timer, err := getHistogramTimer(histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(requestTime[i]*1000) * time.Millisecond)
			//timer指标收集
			timer.ObserveDuration()

			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().UnixNano())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "secondGoroutine"}
			//生成histogram指标的timer
			timer, err := getHistogramTimer(histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(requestTime[i]*1000) * time.Millisecond)
			timer.ObserveDuration()

			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().UnixNano())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	select {}
}

func getHistogramTimer(name, help string, buckets []float64, labelName, labelValue []string) (*prometheus.Timer, error) {
	histogramMetric := &prometheusAOP.HistogramMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	histogramMetric, collectorErr := histogramMetric.GetCollector(name, help, buckets, labelName)
	if collectorErr != nil {
		return nil, collectorErr
	}

	timer, buildTimerErr := histogramMetric.BuildTimer(labelValue)
	if buildTimerErr != nil {
		return nil, buildTimerErr
	}

	return timer, nil
}
