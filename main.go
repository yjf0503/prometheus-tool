package main

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

func init() {
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(prometheusAOP.Registry, promhttp.HandlerOpts{Registry: prometheusAOP.Registry}))
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()
}

func main() {
	defer prometheusAOP.UnregisterCollectors()

	//testHistogramMetric()
	//testSummaryMetric()
	//testCounterMetric()
	//testGaugeMetric()

	go func() {
		name := "request_counter_total"
		help := "test request counter"
		labelName := []string{"path", "memo"}
		testSummaryMetric(name, help, requestTimeObjective, labelName)
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		name := "request_counter_total"
		help := "test request counter"
		labelName := []string{"path", "memo"}
		testSummaryMetric(name, help, requestTimeObjective, labelName)
	}()

	select {}
}

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

var requestTime = []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
var requestTimeBucket = []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000}
var requestTimeObjective = map[float64]float64{0.5: 0.05, 0.8: 0.001, 0.9: 0.01, 0.95: 0.01}

func testHistogramMetric(name, help string, buckets []float64, labelName []string) {
	histogramMetric := &prometheusAOP.HistogramMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	histogramMetric = histogramMetric.CheckAndRegisterCollector(name, help, buckets, labelName)
	for i := 0; i < len(requestTime); i++ {
		//收集histogram指标
		err := histogramMetric.DoObserve([]string{requestApi[i], "test"}, requestTime[i])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func testSummaryMetric(name, help string, objectives map[float64]float64, labelName []string) {
	summaryMetric := &prometheusAOP.SummaryMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	summaryMetric = summaryMetric.CheckAndRegisterCollector(name, help, objectives, labelName)
	for i := 0; i < len(requestTime); i++ {
		//收集summary指标
		err := summaryMetric.DoObserve([]string{requestApi[i], "test"}, requestTime[i])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func testCounterMetric(name, help string, labelName []string) {
	counterMetric := &prometheusAOP.CounterMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	counterMetric = counterMetric.CheckAndRegisterCollector(name, help, labelName)
	for i := 0; i < len(requestTime); i++ {
		//收集counter指标
		api := requestApi[i]
		err := counterMetric.DoObserve([]string{api, "test"}, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", api, time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func testGaugeMetric(name, help string, labelName []string) {
	gaugeMetric := &prometheusAOP.GaugeMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	gaugeMetric = gaugeMetric.CheckAndRegisterCollector(name, help, labelName)
	for i := 0; i < len(requestTime); i++ {
		//收集gauge指标
		api := requestApi[i]
		err := gaugeMetric.DoObserve([]string{api, "test"}, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", api, time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}
