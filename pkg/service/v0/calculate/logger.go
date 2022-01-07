package calculate

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"text/template"
	"time"

	appConfig "owl-engine/pkg/config"
	"owl-engine/pkg/dao/mysql/event"
	"owl-engine/pkg/lib/job"
	"owl-engine/pkg/model/apiModel"
	"owl-engine/pkg/model/dbModel"
	"owl-engine/pkg/util"
	"owl-engine/pkg/xlogs"

	"github.com/elastic/go-elasticsearch/v7"
	uuid "github.com/satori/go.uuid"
)

var (
	LoggerRuleCh    = make(chan map[string]*apiModel.LoggerRule, 1)
	LoggerTaskQueue = make(map[string]string)
)

type loggerRuleCalculate struct {
	Params *apiModel.LoggerRule
}

// Logger 日志计算匹配规则
func Logger(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// 查询数据库, 加载日志处理规则
	condition := apiModel.LoggerRuleCondition{
		Switch: 1, // 是否启用, 1 --- on; 2 --- off
		Inuse:  2, // 是否删除, 1 --- yes; 2 --- no
	}

	loggerRule := new(dbModel.LoggerRule)
	records, _, err := loggerRule.Select(&condition)
	if err != nil {
		xlogs.Errorf("query logger rules from table %s error, %s", loggerRule.TableName(), err.Error())
		return
	}

	cronTab := job.NewCronTab()

	if len(*records) > 0 {
		for _, v := range *records {
			groups := make([]int, 0)
			for _, group := range strings.Split(v.GroupIp, ",") {
				groups = append(groups, util.StringToInt(group))
			}

			calculate := new(loggerRuleCalculate)
			calculate.Params = &apiModel.LoggerRule{
				Id:                v.ID,
				Name:              v.Name,
				Source:            v.Source,
				Address:           v.Address,
				Username:          v.Username,
				Password:          v.Password,
				Index:             v.Index,
				MessageField:      v.MessageField,
				Sql:               v.Sql,
				Threshold:         v.Threshold,
				Origin:            v.Origin,
				BusinessType:      v.BusinessType,
				Category:          v.Category,
				Level:             v.Level,
				Creator:           v.Creator,
				Updater:           v.Updater,
				ResponsiblePeople: v.ResponsiblePeople,
				Crontab:           v.Crontab,
				Switch:            v.Switch,
				Inuse:             v.Inuse,
				GroupId:           groups,
				Description:       v.Description,
				CreatedAt:         util.DateTimeToString(v.CreatedAt),
				UpdatedAt:         util.DateTimeToString(v.UpdatedAt),
			}

			// 添加到定时任务
			id := uuid.NewV4().String()
			if err := cronTab.AddByID(id, v.Crontab, calculate); err == nil {
				LoggerTaskQueue[v.Name] = id
			} else {
				jsonStr, _ := json.Marshal(v)
				xlogs.Errorf(fmt.Sprintf("add cron task for logger rule [%s] error: %s", string(jsonStr), err.Error()))
			}

			time.Sleep(1 * time.Millisecond)
		}

		xlogs.Infof("successfully loaded %d logger rules", len(*records))
	}

	cronTab.Start()

	// 加入定时任务
	for {
		select {
		case value, ok := <-LoggerRuleCh:
			if ok {
				reValue := reflect.ValueOf(value)

				for _, key := range reValue.MapKeys() {
					k := key.String()
					syncLogger(k, value[k], cronTab)
				}
			}
		case <-stopCh:
			ids := cronTab.IDs()
			for _, id := range ids {
				cronTab.DelByID(id)
			}
			cronTab.Stop()

			xlogs.Info("logger expression rule timed task stop calculation")
			return
		}
	}
}

// Run 计算
// 设计思路: 为灵活给予用户进行配置, 采用短链接进行 cron 的定时查询es数据
// golang 中的短链接编程注意事项:
// 陷阱一: Response body 没有及时关闭
// 		网络程序运行中,过了一段时间,比较常见的问题就是爆出错误："socket: too many open files",
//		这通常是由于打开的文件句柄没有关闭造成的。在http使用中,最容易让人忽视的, 就是http返回的response的body必须close,否则就会有内存泄露。
//		更不容易发现的问题是,如果response.body的内容没有被ioutil.ReadAll正确读出来, 也会造成socket链接泄露,后续的服务无法使用。
//		这里, response.body 是一个io.ReadCloser类型的接口， 包含了read和close接口。
func (l *loggerRuleCalculate) Run() {
	switch l.Params.Source {
	case "es":
		params := l.Params

		conf := appConfig.Get()
		// 测试环境使用代理
		var config elasticsearch.Config
		if conf.ServerOptions.EnableProxy && strings.Compare(conf.ServerOptions.Proxy, "") != 0 {
			proxyUrl, _ := url.Parse(conf.ServerOptions.Proxy)
			config = elasticsearch.Config{
				Addresses:  strings.Split(params.Address, ","),
				Username:   params.Username,
				Password:   params.Password,
				MaxRetries: 3,
				Transport: &http.Transport{
					MaxIdleConnsPerHost:   10,
					ResponseHeaderTimeout: 5 * time.Second,
					DialContext: (&net.Dialer{
						Timeout: 5 * time.Second,
					}).DialContext,
					Proxy:             http.ProxyURL(proxyUrl),
					TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS11},
					DisableKeepAlives: true, // 使用短链接进行请求, 注意: http 的 keepalive 和 tcp 的 keepalive 的区别
				},
			}
		} else {
			config = elasticsearch.Config{
				Addresses:  strings.Split(params.Address, ","),
				Username:   params.Username,
				Password:   params.Password,
				MaxRetries: 3,
				Transport: &http.Transport{
					MaxIdleConnsPerHost:   10,
					ResponseHeaderTimeout: 5 * time.Second,
					DialContext: (&net.Dialer{
						Timeout: 5 * time.Second,
					}).DialContext,
					TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS11},
					DisableKeepAlives: true, // 使用短链接进行请求, 注意: http 的 keepalive 和 tcp 的 keepalive 的区别
				},
			}
		}

		var err error
		es, err := elasticsearch.NewClient(config)
		if err == nil {
			// 将 sql string 转换为 map[string]interface
			var query map[string]interface{}
			_ = json.Unmarshal([]byte(params.Sql), &query)

			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(query); err == nil {
				response, err := es.Search(es.Search.WithContext(context.Background()),
					es.Search.WithIndex(params.Index),
					es.Search.WithBody(&buf),
					es.Search.WithTrackTotalHits(true),
					es.Search.WithPretty())

				// 注意！注意！注意！要及时关闭
				defer response.Body.Close()

				if err == nil {
					var result map[string]interface{}
					if err := json.NewDecoder(response.Body).Decode(&result); err == nil {
						// 注意: 在填写 es sql 查询时, 需要经过校验, 通过后才会加载到定时任务; 避免因为语句错误导致返回结果中没有 .hits.total.value 不存在
						count := result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)
						if count >= params.Threshold { // 触发告警
							messages := make([]string, 0)
							hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
							if len(hits) > 0 {
								// 需要对于相似度较高的内容进行聚合
								for _, hit := range hits {
									messages = append(messages, "{"+hit.(map[string]interface{})["_source"].(map[string]interface{})[params.MessageField].(string)+"}")
								}
							}
							message := strings.Replace(strings.Trim(fmt.Sprint(messages), "[]"), " ", " ", -1)
							l.warning(count, message, params)
						}
					}
				} else {
					xlogs.Error(fmt.Sprintf("elasticsearch query sql: %s to response error: %s", params.Sql, err.Error()))
				}
			} else {
				xlogs.Error(fmt.Sprintf("elasticsearch query sql: %s error: %s", params.Sql, err.Error()))
			}
		} else {
			xlogs.Error(fmt.Sprintf("creating the elasticsearch client error: %s", err.Error()))
		}
	default:
		xlogs.Errorf("not realize the calculation for %s", l.Params.Source)
	}
}

// 监听规则变化
func syncLogger(action string, record *apiModel.LoggerRule, cron *job.CronTab) {
	switch strings.ToLower(action) {
	case "delete":
		id := LoggerTaskQueue[record.Name]
		cron.DelByID(id)
		delete(MathTaskQueue, record.Name)

		xlogs.Infof("logger rule [name = %s] has been deleted", record.Name)
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
			var calculate = new(loggerRuleCalculate)
			calculate.Params = record // 参数传递

			if err := cron.AddByID(id, record.Crontab, calculate); err == nil {
				LoggerTaskQueue[record.Name] = id
			} else {
				cron.Stop()

				jsonStr, _ := json.Marshal(record)
				xlogs.Infof(fmt.Sprintf("add cron task for logger rule [%s] error: %s", string(jsonStr), err.Error()))
			}
		}

		xlogs.Infof("logger rule [name = %s] has been updated or added", record.Name)
	}
}

// 发送告警
func (l *loggerRuleCalculate) warning(calValue float64, message string, data *apiModel.LoggerRule) {
	// 告警记录插入数据库
	var content = fmt.Sprintf("规则名称 【%s】触发告警, 当前值为: %v, 阈值为: %v", data.Name, calValue, data.Threshold)

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

	var alertTemplate = `
告警名称：{{ .Name }}
告警类型：{{ .Category }}
业务域： {{ .Type }}
告警源：{{ .Origin }}
告警内容：{{ .Content }}
告警详情: {{ .Message }}
告警值：{{ .Value }}
告警时间：{{ .Datetime }}
负责人：{{ .ResponsiblePeople }}
`
	var params = struct {
		Name              string  `json:"name"`
		Type              string  `json:"type"`
		Category          string  `json:"category"`
		Origin            string  `json:"origin"`
		Content           string  `json:"content"`
		Message           string  `json:"message"`
		Value             float64 `json:"value"`
		Datetime          string  `json:"datetime"`
		ResponsiblePeople string  `json:"responsible_people"`
	}{
		Name:              data.Name,
		Type:              data.BusinessType,
		Category:          category,
		Origin:            data.Origin,
		Content:           content,
		Message:           message,
		Value:             calValue,
		Datetime:          util.DateTimeToString(time.Now()),
		ResponsiblePeople: data.ResponsiblePeople,
	}

	result, _ := template.New("test").Parse(alertTemplate)
	var buffer bytes.Buffer

	err := result.Execute(&buffer, params)
	if err != nil {
		xlogs.Error(fmt.Sprintf("template alert event error: %s", err.Error()))
		return
	}

	alertId := uuid.NewV4().String()

	var record = dbModel.Alert{
		AlertId:      alertId,
		Name:         data.Name,
		Item:         "",
		Origin:       data.Origin,
		BusinessType: data.BusinessType,
		Category:     data.Category,
		Value:        calValue,
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
	conf := appConfig.Get()
	if len(conf.EventOptions.Hooks) > 0 {
		for _, hook := range conf.EventOptions.Hooks {
			if msg, err := Post(hook, alert); err == nil {
				xlogs.Infof("post request to [%s] for alert id [%s] success, response result: [%v]", hook, alert.UUID, msg)
			} else {
				xlogs.Errorf("post data [%s] to %s fail, error message: %s", string(jsonStr), err.Error())
			}
		}
	}
}
