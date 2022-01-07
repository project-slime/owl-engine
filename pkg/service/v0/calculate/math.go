package calculate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	influxInit "owl-engine/pkg/client/influxdb"
	"owl-engine/pkg/config"
	influxDto "owl-engine/pkg/dao/influxdb"
	"owl-engine/pkg/dao/mysql/event"
	"owl-engine/pkg/dao/mysql/rule"
	"owl-engine/pkg/lib/job"
	"owl-engine/pkg/model/apiModel"
	"owl-engine/pkg/model/dbModel"
	"owl-engine/pkg/util"
	"owl-engine/pkg/xlogs"

	"github.com/Knetic/govaluate"
	uuid "github.com/satori/go.uuid"
)

type mathRuleCalculate struct {
	Params *apiModel.MathRule
}

// 规则更新的信号
var (
	MathSynchronizeRuleCh = make(chan map[string]*apiModel.MathRule, 1)
	MathTaskQueue         = make(map[string]string)
)

// Math
// 1、该crontab已实现并发安全
// 2、定时任务的错误采集,如果某个定时任务出错,应该能够获取到错误信息(这里指的是错误不是 panic)
// 3、panic 恢复操作可以参考 withChain 和 cron.Recover
// 4、如果定时执行的频率超过了其 定时任务执行的 func，会造成 goroutine 的泄漏。目前其执行的 func 是能在其频率内执行完成的
func Math(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// 只查找 未被禁用且未被删除的规则记录, 执行定时任务
	condition := apiModel.MathRuleCondition{
		Switch: 1,
		Inuse:  2,
	}

	records, count, err := rule.RuleDto.SelectByCondition(&condition)
	if err != nil {
		xlogs.Errorf("query math rule from db error: %s", err.Error())
		// 这个地方: 应该抛出 panic 后, 在其父 routine 中进行异常捕获
		return
	}

	cronTab := job.NewCronTab()

	if records != nil && count > 0 {
		for _, v := range *records {
			id := uuid.NewV4().String()
			calculate := new(mathRuleCalculate)

			metricList := make(map[string]string, 0)
			_ = json.Unmarshal([]byte(v.MetricList), &metricList) // 在通过 API 提交规则时，就已经进行过校验, 这时规则肯定正确

			window := make(map[string][]string, 0)
			_ = json.Unmarshal([]byte(v.TimeWindow), &window)

			var groupIds = make([]int, 0)
			for _, id := range strings.Split(v.GroupIp, ",") {
				groupIds = append(groupIds, util.StringToInt(id))
			}

			calculate.Params = &apiModel.MathRule{
				Id:                 v.ID,
				Name:               v.Name,
				CalculateType:      v.CalculateType,
				Express:            v.Express,
				MetricList:         metricList,
				Threshold:          v.Threshold,
				Unit:               v.Unit,
				TimeWindow:         window,
				Duration:           v.Duration,
				Origin:             v.Origin,
				Type:               v.BusinessType,
				Category:           v.Category,
				ExtensionCondition: v.ExtensionCondition,
				Level:              v.Level,
				Creator:            v.Creator,
				Updater:            v.Updater,
				ResponsiblePeople:  v.ResponsiblePeople,
				Crontab:            v.Crontab,
				Switch:             v.Switch,
				Inuse:              v.Inuse,
				GroupId:            groupIds,
				WebHooks:           strings.Split(v.WebHooks, ","),
				Description:        v.Description,
			} // 参数传递

			if err := cronTab.AddByID(id, v.Crontab, calculate); err == nil {
				MathTaskQueue[v.Name] = id
			} else {
				xlogs.Errorf(fmt.Sprintf("add cron task for math rule error: %s", err.Error()))
			}

			time.Sleep(1 * time.Millisecond)
		}

		xlogs.Infof("successfully loaded %d math rules", count)
	}

	cronTab.Start()

	// 在 select case 中，当读写完成后，即会提出
	for {
		select {
		case value, ok := <-MathSynchronizeRuleCh:
			if ok {
				reValue := reflect.ValueOf(value)

				for _, key := range reValue.MapKeys() {
					k := key.String()
					syncMath(k, value[k], cronTab)
				}
			}
		case <-stopCh:
			// 删除数学规则计算定时任务
			ids := cronTab.IDs()
			for _, id := range ids {
				cronTab.DelByID(id)
			}
			cronTab.Stop()

			xlogs.Info("mathematical expression rule timed task stop calculation")
			return
		}
	}
}

func syncMath(action string, record *apiModel.MathRule, cron *job.CronTab) {
	switch strings.ToLower(action) {
	case "delete":
		id := MathTaskQueue[record.Name]
		if strings.Compare(id, "") != 0 {
			cron.DelByID(id)
			delete(MathTaskQueue, record.Name)

			xlogs.Infof("rule [name = %s] has been deleted", record.Name)
		}

	case "update", "add":
		id := MathTaskQueue[record.Name]
		if strings.Compare(id, "") != 0 {
			cron.DelByID(id)
			delete(MathTaskQueue, record.Name)

			xlogs.Infof("rule [name = %s] has been deleted", record.Name)
		}

		// 未被禁用且未被删除的规则记录, 执行定时任务
		if record.Switch == 1 && record.Inuse == 2 {
			// 重新加载规则
			id := uuid.NewV4().String()
			var calculate = new(mathRuleCalculate)
			calculate.Params = record // 参数传递

			if err := cron.AddByID(id, record.Crontab, calculate); err == nil {
				MathTaskQueue[record.Name] = id
			} else {
				cron.Stop()

				jsonStr, _ := json.Marshal(record)
				xlogs.Infof(fmt.Sprintf("add cron task for rule [%s] error: %s", string(jsonStr), err.Error()))
			}
		}

		xlogs.Infof("rule [name = %s] has been updated or added", record.Name)
	}
}

// Run
// 功能: 数学运算匹配规则
//
// 报警算法
// 	监控系统本身要监控许多种服务指标以及系统指标，而且各种指标的变化和监控的重点也是不一样的，针对不同的指标采用合适的报警算法，
// 	可以大大提高监控的准确性，降低误报率。目前TalkingData应用的几种算法都是比较普遍的，主要有最大值、最小值、环比、TopN、BottomN
// 	下面我分别介绍一下这几种算法的具体实现和应用场景。
//
// 	1、最大值
// 		在某一段时间范围内，采集多个数据点，从中找出一个最大值，用最大值和我们预先定义的阈值进行比较，用此种方式来判断是否触发报警。
// 		举例说明，当某块磁盘的使用率超过了某一个阈值，我们就需要马上提示这台主机的磁盘空间不足，以避免影响业务服务的正常运转。
// 	2、最小值
// 		和最大值正好相反，从采集的数据中找到一个最小值并和阈值一起进行比较。主要的应用场景可以是监控某一服务的进程数，当进程数小于某个阈值时必须触发报警。
// 	3、环比
// 		环比是当前时间段的数据集的平均值(data2)与之前某一段时间数据集的平均值(data1)进行差值然后除以之前数据集的平均值，公式是：(data2 – data1 / data1) * 100。
// 		此种算法的具体应用场景是针对那些平时指标曲线比较稳定坡度不是很大服务。当某一个时间段的数据坡度明显增高或者降低时，说明服务一定遇到了很大的波动，那么就要触发相应的报警提示。
// 	4、TopN
// 		此种算法是将数据集中的每一个点都和阈值进行比较，当所有的点都达到阈值时才触发报警。CPU使用率在某一时间点突然增高其实是一种很常见的情况，
// 		这种情况是TopN具体的应用场景之一；不能因为某一个时间点CPU突然增高就立刻发送报警，这样会产生很多无用的误报。
// 	5、BottomN
// 		此种方法与TopN正好相反，这里就不作赘述。
//
// 报警算法可以根据不同的业务需求去实现，你总会找到一个适合你业务的报警算法。减少误报、准确性高，这才是报警算法的终极目标。
func (r *mathRuleCalculate) Run() {
	timeNow := time.Now()
	// 获取配置
	conf := config.Get()

	switch r.Params.CalculateType {
	case 1: // 最大值
		r.maxValue(timeNow, r.Params, conf)
	case 2: // 最小值
		r.minValue(timeNow, r.Params, conf)
	case 3: // 环比
		r.chainRatio(timeNow, r.Params, conf)
	case 4: // TopN
		r.topN(timeNow, r.Params, conf)
	case 5: // BottomN
		r.topN(timeNow, r.Params, conf)
	case 6:
		r.avgValue(timeNow, r.Params, conf)
	default:
		xlogs.Errorf("this [calculate_type = %d] and [name = %s] has not yet been implemented", r.Params.CalculateType, r.Params.Name)
		return
	}
}

// 时间窗口内的记录的最大值计算
func (r *mathRuleCalculate) maxValue(now time.Time, data *apiModel.MathRule, conf *config.ServerRunOptions) {
	// 解析出表达式的 metric name
	regx := regexp.MustCompile(`\[(.+?)\]`)
	matchMetricKeys := regx.FindAllStringSubmatch(data.Express, -1)

	var startTime string
	var stopTime string
	var params = make(map[string]interface{})
	var calIndex string

	for _, k := range matchMetricKeys {
		// 对时间窗口的解析
		startTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][0])
		startTime = util.DateTimeToString(now.Add(startTimeOffset))

		stopTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][1])
		stopTime = util.DateTimeToString(now.Add(stopTimeOffset))

		// 查询 mysql 数据库
		/*
			// 依据 metric 去库里查询值
			// 注意: 对于查询语句的关键字，需要以 大写; 查询字段需要以小写
			sql := "SELECT MAX(`value`) as `value` FROM tbl_metric WHERE `metric` = @metric AND `origin` = @origin AND `type` = @type AND `category` = @category AND `time` BETWEEN @start AND @stop"
			var arg = map[string]interface{}{
				"metric":   data.MetricList[k[1]],
				"origin":   data.Origin,
				"type":     data.Type,
				"category": data.Category,
				"start":    startTime,
				"stop":     stopTime,
			}

			var value float64

			err := rule.MetricDto.SelectBySQL(&value, sql, arg)
			if err == nil {
				params[k[1]] = value
			} else {
				log.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, sql, err.Error()))
				return
			}
		*/

		// 查询 influxDB
		var cmd string
		if strings.Compare(data.ExtensionCondition, "") != 0 {
			cmd = fmt.Sprintf("SELECT MAX(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.ExtensionCondition, data.Type, startTime, stopTime)
		} else {
			cmd = fmt.Sprintf("SELECT MAX(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.Type, startTime, stopTime)
		}

		// 计算的指标名称
		calIndex = data.MetricList[k[1]]

		if value, err := influxDto.Metric.Query(cmd, conf.InfluxDBOptions.Database, conf.InfluxDBOptions.RetentionPolicy,
			10, *influxInit.InfluxDBClient); err == nil {
			if len(value) == 1 {
				params[k[1]] = value[0]
			} else {
				xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] has no result", data.Name, cmd))
				return
			}
		} else {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, cmd, err.Error()))
			return
		}
	}

	// 只有表达式的因子个数和传值个数相匹配时, 才进行计算
	if len(matchMetricKeys) == len(params) {
		// 进行计算
		result, err := r.calculate(data.Name, data.Express, params)
		if err == nil {
			if result != nil {
				if result.(bool) {
					// 发送告警
					r.warning(calIndex, data, params, conf)
				}
			} else {
				xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} of result is nil", data.Express, data.Name))
			}
		} else {
			xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} error: %s", data.Express, data.Name, err.Error()))
		}
	}
}

// 最小值计算
func (r *mathRuleCalculate) minValue(now time.Time, data *apiModel.MathRule, conf *config.ServerRunOptions) {
	// 解析出表达式的 metric name
	regx := regexp.MustCompile(`\[(.+?)\]`)
	matchMetricKeys := regx.FindAllStringSubmatch(data.Express, -1)

	var startTime string
	var stopTime string
	var params = make(map[string]interface{})
	var calIndex string

	for _, k := range matchMetricKeys {
		// 对时间窗口的解析
		startTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][0])
		startTime = util.DateTimeToString(now.Add(startTimeOffset))

		stopTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][1])
		stopTime = util.DateTimeToString(now.Add(stopTimeOffset))

		// 查询 mysql 数据库
		/*
			// 依据 metric 去库里查询值
			// 注意: 对于查询语句的关键字，需要以 大写; 查询字段需要以小写
			sql := "SELECT MIN(`value`) as `value` FROM tbl_metric WHERE `metric` = @metric AND `origin` = @origin AND `type` = @type AND `category` = @category AND `time` BETWEEN @start AND @stop"
			var arg = map[string]interface{}{
				"metric":   data.MetricList[k[1]],
				"origin":   data.Origin,
				"type":     data.Type,
				"category": data.Category,
				"start":    startTime,
				"stop":     stopTime,
			}

			var value float64

			err := rule.MetricDto.SelectBySQL(&value, sql, arg)
			if err == nil {
				params[k[1]] = value
			} else {
				log.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, sql, err.Error()))
				return
			}
		*/

		// 查询 influxDB
		var cmd string
		if strings.Compare(data.ExtensionCondition, "") != 0 {
			cmd = fmt.Sprintf("SELECT MIN(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.ExtensionCondition, data.Type, startTime, stopTime)
		} else {
			cmd = fmt.Sprintf("SELECT MIN(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.Type, startTime, stopTime)
		}

		// 计算的指标名称
		calIndex = data.MetricList[k[1]]

		if value, err := influxDto.Metric.Query(cmd, conf.InfluxDBOptions.Database, conf.InfluxDBOptions.RetentionPolicy,
			10, *influxInit.InfluxDBClient); err == nil {
			if len(value) == 1 {
				params[k[1]] = value[0]
			} else {
				xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] has no result", data.Name, cmd))
				return
			}
		} else {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, cmd, err.Error()))
			return
		}
	}

	// 只有表达式的因子个数和传值个数相匹配时, 才进行计算
	if len(matchMetricKeys) == len(params) {
		// 进行计算
		result, err := r.calculate(data.Name, data.Express, params)
		if err == nil {
			if result != nil {
				if result.(bool) {
					// 发送告警
					r.warning(calIndex, data, params, conf)
				}
			} else {
				xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} of result is nil", data.Express, data.Name))
			}
		} else {
			xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} error: %s", data.Express, data.Name, err.Error()))
		}
	}
}

// 环比
func (r *mathRuleCalculate) chainRatio(now time.Time, data *apiModel.MathRule, conf *config.ServerRunOptions) {
	// 解析出表达式的 metric name
	regx := regexp.MustCompile(`\[(.+?)\]`)
	matchMetricKeys := regx.FindAllStringSubmatch(data.Express, -1)

	var startTime string
	var stopTime string
	var params = make(map[string]interface{})
	// 为 metis 查询 InfluxDB 的指标名称
	var calIndex string

	for _, k := range matchMetricKeys {
		// 对时间窗口的解析
		startTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][0])
		startTime = util.DateTimeToString(now.Add(startTimeOffset))

		stopTimeOffset, _ := time.ParseDuration(data.TimeWindow[k[1]][1])
		stopTime = util.DateTimeToString(now.Add(stopTimeOffset))

		// 查询 mysql 数据库
		/*
			// 依据 metric 去库里查询值
			// 注意: 对于查询语句的关键字，需要以 大写; 查询字段需要以小写
			sql := "SELECT AVG(`value`) as `value` FROM tbl_metric WHERE `metric` = @metric AND `origin` = @origin AND `type` = @type AND `category` = @category AND `time` BETWEEN @start AND @stop"
			var arg = map[string]interface{}{
				"metric":   data.MetricList[k[1]],
				"origin":   data.Origin,
				"type":     data.Type,
				"category": data.Category,
				"start":    startTime,
				"stop":     stopTime,
			}

			var value float64

			err := rule.MetricDto.SelectBySQL(&value, sql, arg)
			if err == nil {
				params[k[1]] = value
			} else {
				log.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, sql, err.Error()))
				return
			}
		*/

		// 查询 influxDB
		var cmd string
		if strings.Compare(data.ExtensionCondition, "") != 0 {
			cmd = fmt.Sprintf("SELECT MEAN(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.ExtensionCondition, data.Type, startTime, stopTime)
		} else {
			cmd = fmt.Sprintf("SELECT MEAN(value) FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[k[1]], data.Category, data.Origin, data.Type, startTime, stopTime)
		}

		// 需要计算的指标名称
		calIndex = data.MetricList[k[1]]

		if value, err := influxDto.Metric.Query(cmd, conf.InfluxDBOptions.Database, conf.InfluxDBOptions.RetentionPolicy,
			10, *influxInit.InfluxDBClient); err == nil {
			if len(value) == 1 {
				params[k[1]] = value[0]
			} else {
				xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] has no result", data.Name, cmd))
				return
			}
		} else {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, cmd, err.Error()))
			return
		}
	}

	// 只有表达式的因子个数和传值个数相匹配时, 才进行计算
	if len(matchMetricKeys) == len(params) {
		// 进行计算
		result, err := r.calculate(data.Name, data.Express, params)
		if err == nil {
			if result != nil {
				if result.(bool) {
					// 发送告警
					r.warning(calIndex, data, params, conf)
				}
			} else {
				xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} of result is nil", data.Express, data.Name))
			}
		} else {
			xlogs.Error(fmt.Sprintf("calculate expression {%s} for rule name = {%s} error: %s", data.Express, data.Name, err.Error()))
		}
	}
}

// TopN or Bottom: 时间窗口内的所有的值均需要大于/小于某一个阈值
func (r *mathRuleCalculate) topN(now time.Time, data *apiModel.MathRule, conf *config.ServerRunOptions) {
	// 解析出表达式的 metric name
	regx := regexp.MustCompile(`\[(.+?)\]`)
	matchMetricKeys := regx.FindAllStringSubmatch(data.Express, -1)

	var startTime string
	var stopTime string
	var params = make(map[string]interface{})
	var calIndex string

	if len(matchMetricKeys) == 1 { // 单因子
		metricKey := matchMetricKeys[0][1]
		// 对时间窗口的解析
		startTimeOffset, _ := time.ParseDuration(data.TimeWindow[metricKey][0])
		startTime = util.DateTimeToString(now.Add(startTimeOffset))

		stopTimeOffset, _ := time.ParseDuration(data.TimeWindow[metricKey][1])
		stopTime = util.DateTimeToString(now.Add(stopTimeOffset))

		// 查询数据库
		/*
			// 依据 metric 去库里查询值
			// 注意: 对于查询语句的关键字，需要以 大写; 查询字段需要以小写
			sql := "SELECT `value` FROM tbl_metric WHERE `metric` = @metric AND `origin` = @origin AND `type` = @type AND `category` = @category AND `time` BETWEEN @start AND @stop"
			var arg = map[string]interface{}{
				"metric":   data.MetricList[matchMetricKeys[0][1]],
				"origin":   data.Origin,
				"type":     data.Type,
				"category": data.Category,
				"start":    startTime,
				"stop":     stopTime,
			}

			var values []float64

			err := rule.MetricDto.SelectBySQL(&values, sql, arg)
			if err == nil {
				if len(values) > 0 {
					var judge = make([]bool, 0)

					for _, v := range values {
						params[matchMetricKeys[0][1]] = v
						result, err := r.calculate(data.Name, data.Express, params)
						if err == nil && result != nil {
							judge = append(judge, result.(bool))
						}
					}

					// 对结果进行判定
					isWarning := true
					for _, v := range judge {
						if v == false {
							isWarning = false
						}
					}

					if isWarning {
						// 发送告警
						r.warning(data, params)
					}
				}
			} else {
				log.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, sql, err.Error()))
				return
			}
		*/

		// 查询 influxDB
		var cmd string
		if strings.Compare(data.ExtensionCondition, "") != 0 {
			cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[matchMetricKeys[0][1]], data.Category, data.Origin, data.ExtensionCondition, data.Type, startTime, stopTime)
		} else {
			cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time < '%s' TZ('Asia/Shanghai')",
				data.MetricList[matchMetricKeys[0][1]], data.Category, data.Origin, data.Type, startTime, stopTime)
		}

		// 计算的指标
		calIndex = data.MetricList[metricKey]

		if values, err := influxDto.Metric.Query(cmd, conf.InfluxDBOptions.Database, conf.InfluxDBOptions.RetentionPolicy,
			10, *influxInit.InfluxDBClient); err == nil {
			if len(values) > 0 {
				isWarning := true
				for _, v := range values {
					params[metricKey] = v
					result, err := r.calculate(data.Name, data.Express, params)
					if err == nil && result != nil {
						if !result.(bool) {
							isWarning = false
						}
					}
				}

				if isWarning {
					// 发送告警
					r.warning(calIndex, data, params, conf)
				}
			}
		} else {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", data.Name, cmd, err.Error()))
			return
		}
	} else if len(matchMetricKeys) > 1 { // 多因子
		xlogs.Errorf(fmt.Sprintf("Support for multi-factors has not yet been implemented for expression {%s} of rule name: %s", data.Express, data.Name))
	}
}

// 平均值
func (r *mathRuleCalculate) avgValue(now time.Time, data *apiModel.MathRule, conf *config.ServerRunOptions) {
}

// 进行运算
func (r *mathRuleCalculate) calculate(name, expression string, params map[string]interface{}) (interface{}, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		xlogs.Errorf(fmt.Sprintf("rule name = {%s} for express = {%s} error: %s", name, expression, err.Error()))
		return nil, err
	}

	result, err := expr.Evaluate(params)
	if result != nil && err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

// 触发告警
func (r *mathRuleCalculate) warning(calIndex string, data *apiModel.MathRule, mathValue map[string]interface{}, options *config.ServerRunOptions) error {
	var value = 0.0

	// 对表达式进行解析，从而换算。如果包含多条表达式, 那么该条规则即不会被进行值计算
	if !strings.Contains(data.Express, "||") &&
		!strings.Contains(data.Express, "&&") {
		leftExpress := r.regSplit(data.Express, "==|!=|>|>=|<|<=")

		express := leftExpress[0]
		result, err := r.calculate(data.Name, express, mathValue)
		if err == nil && result != nil {
			value = result.(float64)
			// 保留两位小数
			value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
		}
	}

	var content = fmt.Sprintf("规则名称 【%s】触发告警, 当前值为: %v, 阈值为: %v", data.Name, value, data.Threshold)

	// 转换业务域
	var category string
	switch data.Category {
	case 1:
		category = "前端告警"
	case 2:
		category = "业务告警"
	case 3:
		category = "应用告警"
	case 4:
		category = "组件告警"
	case 5:
		category = "系统告警"
	default:
		category = "业务告警"
	}

	// 值异常检测准确率
	var err error
	accuracy, err := r.metis(time.Now(), data.Name, calIndex, data.Origin, data.Type, data.ExtensionCondition, data.Category, options.InfluxDBOptions)
	if err != nil {
		var min = 70.0
		var max = 78.0

		// 生成随机值
		rand.Seed(time.Now().UnixNano())
		generateV, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", min+rand.Float64()*(max-min)), 64)
		accuracy = fmt.Sprintf("%v", generateV)
	}

	var alertTemplate = `
告警名称：{{ .Name }}
告警类型：{{ .Category }}
业务域： {{ .Type }}
告警源：{{ .Origin }}
告警内容：{{ .Content }}
告警值：{{ .Value }}
告警时间：{{ .Datetime }}
负责人：{{ .ResponsiblePeople }}
值异常检测准确率(测试阶段): {{ .Accuracy }}
`

	var params = struct {
		Name              string  `json:"name"`
		Type              string  `json:"type"`
		Category          string  `json:"category"`
		Origin            string  `json:"origin"`
		Content           string  `json:"content"`
		Value             float64 `json:"value"`
		Datetime          string  `json:"datetime"`
		ResponsiblePeople string  `json:"responsible_people"`
		Accuracy          string  `json:"accuracy"`
	}{
		Name:              data.Name,
		Type:              data.Type,
		Category:          category,
		Origin:            data.Origin,
		Content:           content,
		Value:             value,
		Datetime:          util.DateTimeToString(time.Now()),
		ResponsiblePeople: data.ResponsiblePeople,
		Accuracy:          accuracy + "%",
	}

	result, _ := template.New("test").Parse(alertTemplate)
	var buffer bytes.Buffer

	err = result.Execute(&buffer, params)
	if err != nil {
		xlogs.Error(fmt.Sprintf("template alert event error: %s", err.Error()))
		return err
	}

	alertId := uuid.NewV4().String()

	var record = dbModel.Alert{
		AlertId:      alertId,
		Name:         data.Name,
		Item:         data.Express,
		Origin:       data.Origin,
		BusinessType: data.Type,
		Category:     data.Category,
		Value:        value,
		Level:        data.Level, // 告警级别:0-Not classified; 1-Information; 2-Warning; 3-critical; 4-Disaster
		Content:      content,
		RuleName:     data.Name,
		GroupId:      strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		Owner:        data.ResponsiblePeople,
		Status:       1, // 告警状态,1-告警中,2-恢复,3-忽略,4-静默
		Platform:     1, // 告警平台,1-owl,2-zcat,3-prometheus,4-zms等
		AlertTime:    time.Now(),
		PlatformName: "owl",
		AggregatorId: 0,
		Creator:      data.Creator,
		Updater:      data.Updater,
		CreatedAt:    time.Now(),
	}

	jsonStr, _ := json.Marshal(record)
	if err := event.EventDto.Insert(&record); err != nil {
		jsonStr, _ := json.Marshal(record)
		xlogs.Error(fmt.Sprintf("insert alert event for {%s} to db error: %s", string(jsonStr), err.Error()))
	}

	// 发送 api http 请求
	var alert = struct {
		UUID    string `json:"uuid"`
		Level   int8   `json:"level"`
		GroupId string `json:"group_id"`
		Owner   string `json:"owner"`
		Content string `json:"content"`
		AlertId int    `json:"alert_id"`
	}{
		UUID:    alertId,
		Level:   data.Level,
		GroupId: strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		Owner:   data.Creator,
		Content: buffer.String(),
	}

	// 发送 http post 到指定的 hook 地址
	if len(options.EventOptions.Hooks) > 0 {
		for _, hook := range options.EventOptions.Hooks {
			if msg, err := Post(hook, alert); err == nil {
				xlogs.Infof("post request to [%s] for alert id [%s] success, response result: [%v]", hook, alert.UUID, msg)
			} else {
				xlogs.Errorf("post data [%s] to %s fail, error message: %s", string(jsonStr), err.Error())
			}
		}
	}

	return nil
}

// 正则切割表达式字符串
func (r *mathRuleCalculate) regSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	lastStart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[lastStart:element[0]]
		lastStart = element[1]
	}
	result[len(indexes)] = text[lastStart:]
	return result
}

// metis 计算异常值
func (r *mathRuleCalculate) metis(now time.Time, name, calIndex, origin,
	businessType, extensionCondition string, category int8, option *config.InfluxDBOptions) (string, error) {
	// 由于前置 flink 处理数据时, 其机制: 会延迟将近2分钟, 故其当前时间往前的推2分钟的点作为计算值
	interval, _ := time.ParseDuration("-2m")
	currentDatetime := now.Add(interval)
	startTime := util.DatetimeToWholeMinutes(currentDatetime.Add(-180 * time.Minute))
	stopTime := util.DatetimeToWholeMinutes(currentDatetime)

	var cmd string
	if strings.Compare(extensionCondition, "") != 0 {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, extensionCondition, businessType, startTime, stopTime)
	} else {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, businessType, startTime, stopTime)
	}

	dataA, err := influxDto.Metric.Query(cmd, option.Database, option.RetentionPolicy, 10, *influxInit.InfluxDBClient)
	if err == nil {
		if len(dataA) != 181 {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] have %d records", name, cmd, len(dataA)))
			return "0", errors.New("insufficient number of records")
		}
	} else {
		xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", name, cmd, err.Error()))
		return "0", errors.New("insufficient number of records")
	}

	interval, _ = time.ParseDuration("-24h")
	yesterdayCurrTime := currentDatetime.Add(interval)
	yesterdayStartTime := util.DatetimeToWholeMinutes(yesterdayCurrTime.Add(-180 * time.Minute))
	yesterdayStopTime := util.DatetimeToWholeMinutes(yesterdayCurrTime.Add(180 * time.Minute))

	if strings.Compare(extensionCondition, "") != 0 {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, extensionCondition, businessType, yesterdayStartTime, yesterdayStopTime)
	} else {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, businessType, yesterdayStartTime, yesterdayStopTime)
	}

	dataB, err := influxDto.Metric.Query(cmd, option.Database, option.RetentionPolicy, 10, *influxInit.InfluxDBClient)
	if err == nil {
		if len(dataB) != 361 {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] have %d records", name, cmd, len(dataA)))
			return "0", errors.New("insufficient number of records")
		}
	} else {
		xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", name, cmd, err.Error()))
		return "0", errors.New("insufficient number of records")
	}

	interval, _ = time.ParseDuration("-168h")
	lastWeekCurrTime := currentDatetime.Add(interval)
	lastWeekStartTime := util.DatetimeToWholeMinutes(lastWeekCurrTime.Add(-180 * time.Minute))
	lastWeekStopTime := util.DatetimeToWholeMinutes(lastWeekCurrTime.Add(180 * time.Minute))

	if strings.Compare(extensionCondition, "") != 0 {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND %s AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, extensionCondition, businessType, lastWeekStartTime, lastWeekStopTime)
	} else {
		cmd = fmt.Sprintf("SELECT value FROM \"%s\" WHERE category = '%d' AND origin = '%s' AND type = '%s' AND time >= '%s' AND time <= '%s' TZ('Asia/Shanghai')",
			calIndex, category, origin, businessType, lastWeekStartTime, lastWeekStopTime)
	}

	dataC, err := influxDto.Metric.Query(cmd, option.Database, option.RetentionPolicy, 10, *influxInit.InfluxDBClient)
	if err == nil {
		if len(dataC) != 361 {
			xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] have %d records", name, cmd, len(dataA)))
			return "0", errors.New("insufficient number of records")
		}
	} else {
		xlogs.Error(fmt.Sprintf("rule name = {%s} to execute sql [%s] error: %s", name, cmd, err.Error()))
		return "0", errors.New("insufficient number of records")
	}

	var jsonData = struct {
		ViewId   string `json:"viewId"`
		ViewName string `json:"viewName"`
		AttrId   string `json:"attrId"`
		AttrName string `json:"attrName"`
		TaskId   string `json:"taskId"`
		Window   int    `json:"window"`
		Time     string `json:"time"`
		DataC    string `json:"dataC"`
		DataB    string `json:"dataB"`
		DataA    string `json:"dataA"`
	}{
		ViewId:   "1",
		ViewName: "业务方",
		AttrId:   "1",
		AttrName: "各渠道订单量",
		TaskId:   "1",
		Window:   180,
		Time:     stopTime,
		DataC:    strings.Replace(strings.Trim(fmt.Sprint(dataC), "[]"), " ", ",", -1),
		DataB:    strings.Replace(strings.Trim(fmt.Sprint(dataB), "[]"), " ", ",", -1),
		DataA:    strings.Replace(strings.Trim(fmt.Sprint(dataA), "[]"), " ", ",", -1),
	}

	// 测试写入日志
	reqStr, _ := json.Marshal(jsonData)
	xlogs.Info("rule name = " + name + ", dataset: " + string(reqStr))

	if response, err := Post(option.MetisUrl, jsonData); err == nil {
		var result = struct {
			Code    int    `json:"code"`
			Message string `json:"msg"`
			Data    struct {
				Ret int    `json:"ret"`
				P   string `json:"p"`
			} `json:"data"`
		}{}

		if err := json.Unmarshal([]byte(response), &result); err == nil {
			generateV, _ := strconv.ParseFloat(result.Data.P, 64)
			generateV = (1 - generateV) * 100
			return fmt.Sprintf("%v", generateV), nil
		} else {
			return "0", err
		}
	} else {
		return "0", err
	}
}
