package test

import (
	"awesomeProject/tools/prometheusAOP"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"strconv"
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
	gaugeMetric, collectorErr := gaugeMetric.GetGaugeVecCollector(name, help, labelName)
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

func TestGaugeTimerMetric(*testing.T) {
	go func() {
		labelName := []string{"path", "memo", "requestID"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine", uniqueId()}

			//生成gauge指标的timer
			timer, err := prometheusAOP.GetGaugeTimer(gaugeMetricName, gaugeMetricHelp, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			//timer指标收集
			timer.ObserveDuration()
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		labelName := []string{"path", "memo", "requestID"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "secondGoroutine", uniqueId()}

			//生成gauge指标的timer
			timer, err := prometheusAOP.GetGaugeTimer(gaugeMetricName, gaugeMetricHelp, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			//timer指标收集
			timer.ObserveDuration()
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	select {}
}

// GetSha256String 生成32位md5字串
func GetSha256String(s string) string {
	h := sha256.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		fmt.Printf("can't generate sha256 string: %v, %v", s, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// UniqueId 生成32位唯一id字串
func uniqueId() string {
	b := make([]byte, 42)
	if _, err := io.ReadFull(cryptoRand.Reader, b); err != nil {
		fmt.Printf("io read error: %v", err)
		return ""
	}

	s := strconv.FormatInt(time.Now().Unix(), 10) + GetSha256String(base64.URLEncoding.EncodeToString(b))[:32]
	return s[:32]
}
