package test

import (
	"awesomeProject/tools/prometheusAOP"
	"fmt"
	"testing"
	"time"
)

var requestTime = []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
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

func testCounterMetric(*testing.T) {

	counterMetric := prometheusAOP.CounterMetric{}
	counterMetric.Before("request_counter_total", "test request counter", []string{"path"})

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
