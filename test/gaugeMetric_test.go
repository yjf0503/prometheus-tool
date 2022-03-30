package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"testing"
	"time"
)

var gaugeMetricName = "request_gauge_total"
var gaugeMetricHelp = "test request gauge"

func TestGaugeMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//收集指标
			err := doGaugeObserve(gaugeMetricName, gaugeMetricHelp, labelName, labelValue, 1)
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
			err := doGaugeObserve(gaugeMetricName, gaugeMetricHelp, labelName, labelValue, 1)
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

func doGaugeObserve(name, help string, labelName, labelValue []string, metricValue float64) error {
	gaugeMetric := &prometheusAOP.GaugeMetric{}
	//通过单例模式获取collector，如果不存在该collector，进行注册并返回
	gaugeMetric, collectorErr := gaugeMetric.GetCollector(name, help, labelName)
	if collectorErr != nil {
		return collectorErr
	}

	//执行指标数据收集
	observeErr := gaugeMetric.DoObserve(labelValue, metricValue)
	if observeErr != nil {
		return observeErr
	}

	return nil
}
