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
		TestCounterMetric()
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		TestCounterMetric()
	}()

	select {}
}

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

var requestTime = []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
var requestTimeBucket = []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000}
var requestTimeObjective = map[float64]float64{0.5: 0.05, 0.8: 0.001, 0.9: 0.01, 0.95: 0.01}

func testHistogramMetric() {
	histogramMetric := prometheusAOP.HistogramMetric{}
	histogramMetric.Before("request_time_histogram", "the relationship between api and request time", requestTimeBucket, []string{"path"})

	for i := 0; i < len(requestTime); i++ {
		//收集histogram指标
		err := histogramMetric.DoObserve([]string{requestApi[i]}, requestTime[i])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %f \n", requestApi[i], requestTime[i])
	}
}

func testSummaryMetric() {
	summaryMetric := prometheusAOP.SummaryMetric{}
	summaryMetric.Before("request_time_summary", "the relationship between api and request time", requestTimeObjective, []string{"path"})

	for i := 0; i < len(requestTime); i++ {
		//收集summary指标
		err := summaryMetric.DoObserve([]string{requestApi[i]}, requestTime[i])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %f \n", requestApi[i], requestTime[i])
	}
}

func TestCounterMetric() {
	name := "request_counter_total"
	help := "test request counter"
	labelName := "path"

	counterMetric := &prometheusAOP.CounterMetric{}
	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	counterMetric = counterMetric.CheckAndRegisterCollector(name, help, labelName)

	for i := 0; i < len(requestTime); i++ {
		//收集counter指标
		api := requestApi[i]
		err := counterMetric.DoObserve([]string{api}, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", api, time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func testGaugeMetric() {
	gaugeMetric := prometheusAOP.GaugeMetric{}
	gaugeMetric.Before("request_gauge_total", "test request gauge", []string{"path"})

	for {
		//收集gauge指标
		api := requestApi[time.Now().Unix()%10]
		err := gaugeMetric.DoObserve([]string{api}, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("requestApi - requestTime: %s - %d \n", api, time.Now().Unix())
		time.Sleep(time.Duration(1) * time.Second)
	}
}
