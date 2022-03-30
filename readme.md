### Prometheus接入组件

**目录 (Table of Contents)**

[TOCM]

[TOC]

# 接入方法
##开启http服务，监听metrics端口

![](/Users/jiefuyang/Desktop/截屏2022-03-30 12.03.39.png)


##获取监控指标对象（以counter类型为例）
	1. 确定需要使用的指标类型，然后生成一个该类型的空对象即可            
	2. 调用GetCollector方法，通过单例模式，获取已经配置好name，labelname等项的指标对象    

![](/Users/jiefuyang/Desktop/截屏2022-03-30 11.59.58.png)

##进行指标数据收集
	调用指标对象的DoObserve方法，传入labelValue和metricValue，至此，prometheus已经开始收集该指标的数据  

![](/Users/jiefuyang/Desktop/截屏2022-03-30 12.06.12.png)


# 内部实现
## 指标注册
## 指标配置
## 数据收集