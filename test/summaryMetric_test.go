package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"testing"
	"time"
)

var summaryMetricName = "request_summary_total"
var summaryMetricHelp = "test request summary"
var requestTimeObjective = map[float64]float64{0.5: 0.05, 0.8: 0.001, 0.9: 0.01, 0.95: 0.01}

func TestSummaryMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//收集指标
			err := doSummaryObserve(summaryMetricName, summaryMetricHelp, requestTimeObjective, labelName, labelValue, requestTime[i])
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
			err := doSummaryObserve(summaryMetricName, summaryMetricHelp, requestTimeObjective, labelName, labelValue, requestTime[i])
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

func doSummaryObserve(name, help string, objectives map[float64]float64, labelName, labelValue []string, metricValue float64) error {
	summaryMetric := &prometheusAOP.SummaryMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	summaryMetric, collectorErr := prometheusAOP.GetSummaryCollector(name, help, objectives, labelName)
	if collectorErr != nil {
		return collectorErr
	}

	//执行指标数据收集
	observeErr := summaryMetric.DoObserve(labelValue, metricValue)
	if observeErr != nil {
		return observeErr
	}

	return nil
}

func TestTimerSummaryMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//生成histogram指标的timer
			timer, err := prometheusAOP.GetSummaryTimer(summaryMetricName, summaryMetricHelp, requestTimeObjective, labelName, labelValue)
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
			timer, err := prometheusAOP.GetSummaryTimer(histogramMetricName, histogramMetricHelp, requestTimeObjective, labelName, labelValue)
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
