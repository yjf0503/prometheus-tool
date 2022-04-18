package test

import (
	"awesomeProject/tools/prometheusAOP"
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var histogramMetricName = "request_histogram_total"
var histogramMetricHelp = "test request histogram"
var requestTimeBucket = []float64{0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
var requestTime = []float64{0.1, 0.15, 0.2, 0.23, 0.25, 0.4, 0.5, 0.7, 0.85, 0.9}

// InitJaeger ...
func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg, err := jaegerCfg.FromEnv()
	cfg.ServiceName = service
	tracer, closer, err := cfg.NewTracer(jaegerCfg.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func TestHistogramMetric(*testing.T) {
	ctx := context.Background()
	tracer, closer := InitJaeger("TestHistogramMetric")
	opentracing.InitGlobalTracer(tracer)
	defer closer.Close()

	go func() {
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "firstGoroutine"}
			//收集非时间指标
			firstGoroutineSpan := tracer.StartSpan("firstGoroutine" + strconv.FormatInt(time.Now().UnixNano(), 10))
			doHistogramObserveCtx := opentracing.ContextWithSpan(ctx, firstGoroutineSpan)
			err := doHistogramObserve(doHistogramObserveCtx, histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue, requestTime[i])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)

			firstGoroutineSpan.Finish()
		}
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			labelValue := []string{requestApi[i], "secondGoroutine"}
			//收集非时间指标
			secondGoroutineSpan := tracer.StartSpan("secondGoroutine" + strconv.FormatInt(time.Now().UnixNano(), 10))
			doHistogramObserveCtx := opentracing.ContextWithSpan(ctx, secondGoroutineSpan)
			err := doHistogramObserve(doHistogramObserveCtx, histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue, requestTime[i])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().Unix())
			time.Sleep(time.Duration(1) * time.Second)

			secondGoroutineSpan.Finish()
		}
	}()

	select {}
}

func doHistogramObserve(ctx context.Context, name, help string, buckets []float64, labelName, labelValue []string, metricValue float64) error {
	doHistogramObserveSpan, ctx := opentracing.StartSpanFromContext(ctx, "doHistogramObserve")
	defer doHistogramObserveSpan.Finish()

	//判断collector是否已注册到prometheus的注册表中，通过单例模式控制
	histogramMetric, collectorErr := prometheusAOP.GetHistogramCollector(name, help, buckets, labelName)
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
		for {
			metricName := "TracksByEntitySortedBySimilarity_request_duration"
			metricHelp := "request histogram"
			labelName := []string{"stage"}

			//-------------------阶段0：全局请求时长-------------------
			totalTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"total"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			//-------------------阶段一：入参校验-------------------
			validateTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"validate"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(50)+100) * time.Millisecond)
			validateTimer.ObserveDuration()

			//---------------阶段二：获取entity对应的轨迹-------------------
			fetchTrackListByEntityTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"fetchTrackListByEntity"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
			fetchTrackListByEntityTimer.ObserveDuration()

			//---------------阶段三：根据轨迹获取特征-------------------
			batchGetFeaturesTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"batchGetFeatures"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+400) * time.Millisecond)
			batchGetFeaturesTimer.ObserveDuration()

			//---------------阶段四：对特征进行比较，获取特征相似度-------------------
			batchCompareTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"batchCompare"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
			batchCompareTimer.ObserveDuration()

			//---------------阶段五：根据特征相似度对轨迹进行排序-------------------
			reorderGetTopKTracksWithSimilarityTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"reorderGetTopKTracksWithSimilarity"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
			reorderGetTopKTracksWithSimilarityTimer.ObserveDuration()

			totalTimer.ObserveDuration()

			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
	go func() {
		for {
			metricName := "TracksByEntitySortedBySimilarity_request_duration"
			metricHelp := "request histogram"
			labelName := []string{"stage"}

			//-------------------阶段0：全局请求时长-------------------
			totalTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"total"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			//-------------------阶段一：入参校验-------------------
			validateTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"validate"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(50)+100) * time.Millisecond)
			validateTimer.ObserveDuration()

			//---------------阶段二：获取entity对应的轨迹-------------------
			fetchTrackListByEntityTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"fetchTrackListByEntity"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
			fetchTrackListByEntityTimer.ObserveDuration()

			//---------------阶段三：根据轨迹获取特征-------------------
			batchGetFeaturesTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"batchGetFeatures"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+400) * time.Millisecond)
			batchGetFeaturesTimer.ObserveDuration()

			//---------------阶段四：对特征进行比较，获取特征相似度-------------------
			batchCompareTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"batchCompare"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
			batchCompareTimer.ObserveDuration()

			//---------------阶段五：根据特征相似度对轨迹进行排序-------------------
			reorderGetTopKTracksWithSimilarityTimer, err := prometheusAOP.GetHistogramTimer(metricName, metricHelp, requestTimeBucket, labelName, []string{"reorderGetTopKTracksWithSimilarity"})
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
			reorderGetTopKTracksWithSimilarityTimer.ObserveDuration()

			totalTimer.ObserveDuration()

			time.Sleep(time.Duration(1) * time.Second)
		}
	}()

	select {}
}
