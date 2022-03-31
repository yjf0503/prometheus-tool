package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"testing"
	"time"
)

var counterMetricName = "request_counter_total"
var counterMetricHelp = "test request counter"
var requestApi = [10]string{
	"add_outside_oplog_async",
	"batch_update_entity_cluster",
	"delete_events_data",
	"add_outside_oplog_async",
	"batch_update_entity_cluster",
	"delete_events_data",
	"add_outside_oplog_async",
	"batch_update_entity_cluster",
	"delete_events_data",
	"add_outside_oplog_async"}

func init() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(prometheusAOP.Registry, promhttp.HandlerOpts{Registry: prometheusAOP.Registry}))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		http.Handle("/metric", promhttp.Handler())
		err := http.ListenAndServe(":8082", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()
}

func TestCounterMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//收集指标
			err := doCounterObserve(counterMetricName, counterMetricHelp, labelName, labelValue, 1)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "secondGoroutine"}
			//收集指标
			err := doCounterObserve(counterMetricName, counterMetricHelp, labelName, labelValue, 1)
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

func doCounterObserve(name, help string, labelName, labelValue []string, metricValue float64) error {
	counterMetric := &prometheusAOP.CounterMetric{}
	//通过单例模式获取collector，如果不存在该collector，进行注册并返回
	counterMetric, collectorErr := counterMetric.GetCollector(name, help, labelName)
	if collectorErr != nil {
		return collectorErr
	}

	//执行指标数据收集
	observeErr := counterMetric.DoObserve(labelValue, metricValue)
	if observeErr != nil {
		return observeErr
	}

	return nil
}
