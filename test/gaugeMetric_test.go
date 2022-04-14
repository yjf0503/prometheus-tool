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

var gaugeMetricName = "TracksByEntitySortedBySimilarity_request_duration"
var gaugeMetricHelp = "request gauge"

func TestGaugeTimerMetric(*testing.T) {
	for {
		labelName := []string{"stage"}

		//-------------------阶段0：全局请求时长-------------------
		//获取total监控计时器
		totalMetric, totalTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)

		//-------------------阶段一：入参校验-------------------
		//获取validate监控计时器
		validateMetric, validateTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)
		//模拟程序执行时间
		time.Sleep(time.Duration(rand.Intn(50)+100) * time.Millisecond)
		//validate监控计时器指标收集
		validateMetric.DoObserveTimer([]string{"validate"}, validateTimerStart)

		//---------------阶段二：获取entity对应的轨迹-------------------
		//获取fetchTrackListByEntity监控计时器
		fetchTrackListByEntityMetric, fetchTrackListByEntityTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)
		//模拟程序执行时间
		time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
		//fetchTrackListByEntity监控计时器指标收集
		fetchTrackListByEntityMetric.DoObserveTimer([]string{"fetchTrackListByEntity"}, fetchTrackListByEntityTimerStart)

		//---------------阶段三：根据轨迹获取特征-------------------
		//获取batchGetFeatures监控计时器
		batchGetFeaturesMetric, batchGetFeaturesTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)
		//模拟程序执行时间
		time.Sleep(time.Duration(rand.Intn(200)+400) * time.Millisecond)
		//batchGetFeatures监控计时器指标收集
		batchGetFeaturesMetric.DoObserveTimer([]string{"batchGetFeatures"}, batchGetFeaturesTimerStart)

		//---------------阶段四：对特征进行比较，获取特征相似度-------------------
		//获取batchCompare监控计时器
		batchCompareMetric, batchCompareTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)
		//模拟程序执行时间
		time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)
		//batchCompareMetric监控计时器指标收集
		batchCompareMetric.DoObserveTimer([]string{"batchCompare"}, batchCompareTimerStart)

		//---------------阶段五：根据特征相似度对轨迹进行排序-------------------
		//获取reorder监控计时器
		reorderMetric, reorderTimerStart := prometheusAOP.GetGaugeCollectorAndSetTimer(gaugeMetricName, gaugeMetricHelp, labelName)
		//模拟程序执行时间
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		//reorder监控计时器指标收集
		reorderMetric.DoObserveTimer([]string{"reorder"}, reorderTimerStart)

		//total监控计时器指标收集
		totalMetric.DoObserveTimer([]string{"total"}, totalTimerStart)

		time.Sleep(time.Duration(1) * time.Second)
	}
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
