package test

import (
	"awesomeProject/tools/prometheusAOP"
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
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
			//创建jaeger全局tracer
			ctx := context.Background()
			tracer, closer := InitJaeger("TestTimerHistogramMetric")
			opentracing.InitGlobalTracer(tracer)

			labelValue := []string{requestApi[i], "firstGoroutine"}
			//生成histogram指标的timer
			firstGoroutineSpan := tracer.StartSpan("firstGoroutine" + strconv.FormatInt(time.Now().UnixNano(), 10))
			getHistogramTimerCtx := opentracing.ContextWithSpan(ctx, firstGoroutineSpan)
			timer, err := getHistogramTimer(getHistogramTimerCtx, histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			//模拟程序执行时间
			simulateExecCtx := opentracing.ContextWithSpan(ctx, firstGoroutineSpan)
			simulateExecSpan, ctx := opentracing.StartSpanFromContext(simulateExecCtx, "simulateExec")
			time.Sleep(time.Duration(requestTime[i]*1000) * time.Millisecond)
			simulateExecSpan.Finish()

			//timer指标收集
			ObserveDurationCtx := opentracing.ContextWithSpan(ctx, firstGoroutineSpan)
			ObserveDurationSpan, ctx := opentracing.StartSpanFromContext(ObserveDurationCtx, "ObserveDuration")
			timer.ObserveDuration()
			ObserveDurationSpan.Finish()

			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().UnixNano())
			time.Sleep(time.Duration(1) * time.Second)

			firstGoroutineSpan.Finish()
			err = closer.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}()

	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		labelName := []string{"path", "memo"}
		for i := 0; i < len(requestApi); i++ {
			//创建jaeger全局tracer
			ctx := context.Background()
			tracer, closer := InitJaeger("TestTimerHistogramMetric")
			opentracing.InitGlobalTracer(tracer)

			labelValue := []string{requestApi[i], "secondGoroutine"}
			//生成histogram指标的timer
			secondGoroutineSpan := tracer.StartSpan("secondGoroutine" + strconv.FormatInt(time.Now().UnixNano(), 10))
			getHistogramTimerCtx := opentracing.ContextWithSpan(ctx, secondGoroutineSpan)
			timer, err := getHistogramTimer(getHistogramTimerCtx, histogramMetricName, histogramMetricHelp, requestTimeBucket, labelName, labelValue)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//模拟程序执行时间
			simulateExecCtx := opentracing.ContextWithSpan(ctx, secondGoroutineSpan)
			simulateExecSpan, ctx := opentracing.StartSpanFromContext(simulateExecCtx, "simulateExec")
			time.Sleep(time.Duration(requestTime[i]*1000) * time.Millisecond)
			simulateExecSpan.Finish()

			//timer指标收集
			ObserveDurationCtx := opentracing.ContextWithSpan(ctx, secondGoroutineSpan)
			ObserveDurationSpan, ctx := opentracing.StartSpanFromContext(ObserveDurationCtx, "ObserveDuration")
			timer.ObserveDuration()
			ObserveDurationSpan.Finish()

			fmt.Printf("requestApi - requestTime: %s - %d \n", requestApi[i], time.Now().UnixNano())
			time.Sleep(time.Duration(1) * time.Second)

			secondGoroutineSpan.Finish()
			err = closer.Close()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}()

	select {}
}

func getHistogramTimer(ctx context.Context, name, help string, buckets []float64, labelName, labelValue []string) (*prometheus.Timer, error) {
	doHistogramObserveSpan, ctx := opentracing.StartSpanFromContext(ctx, "getHistogramTimer")
	defer doHistogramObserveSpan.Finish()

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
