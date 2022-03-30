package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"testing"
	"time"
)

var requestApi = [10]string{
	"add_outside_oplog",
	"batch_update_entity",
	"delete_events",
	"add_outside_oplog",
	"batch_update_entity",
	"delete_events",
	"add_outside_oplog",
	"batch_update_entity",
	"delete_events",
	"add_outside_oplog"}

func init() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(prometheusAOP.Registry, promhttp.HandlerOpts{Registry: prometheusAOP.Registry}))
		err := http.ListenAndServe(":8080", nil)
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
			doObserve("request_counter_total", "test request count", labelName, labelValue, 1)
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
			doObserve("request_counter_total", "test request count", labelName, labelValue, 1)
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	select {}
}

func doObserve(name, help string, labelName, labelValue []string, metricValue float64) {
	counterMetric := &prometheusAOP.CounterMetric{}
	//通过单例模式获取collector，如果不存在该collector，进行注册并返回
	counterMetric, collectorErr := counterMetric.GetCollector(name, help, labelName)
	if collectorErr != nil {
		fmt.Println(collectorErr.Error())
		return
	}

	//执行指标数据收集
	observeErr := counterMetric.DoObserve(labelValue, metricValue)
	if observeErr != nil {
		fmt.Println(observeErr.Error())
		return
	}
}
