package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	ruleDto "owl-engine/pkg/dao/mysql/rule"
	"owl-engine/pkg/model/apiModel"
	"owl-engine/pkg/model/dbModel"
	"owl-engine/pkg/service/v0/calculate"
	"owl-engine/pkg/util"

	"github.com/Knetic/govaluate"
	"github.com/robfig/cron/v3"
)

type mathRule struct{}

var MathRuleSrv = new(mathRule)

// CheckRule 规则合法性校验
func (r *mathRule) CheckRule(data *apiModel.MathRule) (bool, error) {
	// 查看数据库记录是否已经存在
	var condition = apiModel.MathRuleCondition{
		Id:                data.Id,
		Name:              data.Name,
		Creator:           data.Creator,
		ResponsiblePeople: data.ResponsiblePeople,
		Origin:            data.Origin,
		Type:              data.Type,
		Category:          data.Category,
		Switch:            data.Switch,
		Inuse:             data.Inuse,
	}
	if _, count, err := ruleDto.RuleDto.SelectByCondition(&condition); err == nil {
		if count > 0 {
			return false, errors.New("rule already exists for rule name " + data.Name)
		}
	}

	// 计算类型值校验
	switch data.CalculateType {
	case 1, 2, 3, 4, 5:
	default:
		return false, errors.New("the parameter calculate_type is set incorrectly. example: 1 -- Max; 2 -- Min; 3 -- chainRatio; 4 -- TopN; 5 -- BottomN")
	}

	// 持续时间的校验: 必须大于等于 1 的正整数
	if data.Duration < 0 {
		return false, errors.New("the duration of the rule must be greater than 0")
	}

	// 告警接收人列表校验: 不能为空
	if len(data.GroupId) == 0 && len(data.WebHooks) == 0 {
		return false, errors.New("web_hooks: " + "at least one item in the alert recipient list cannot be empty")
	}

	// 关于 crontab 的表达式正则校验
	if _, err := cron.ParseStandard(data.Crontab); err != nil {
		return false, errors.New("cron express: " + err.Error())
	}

	// 表达式的正则校验
	rex := regexp.MustCompile(`\[(.+?)\]`)
	metricList := rex.FindAllStringSubmatch(data.Express, -1)

	if len(metricList) <= 0 {
		return false, errors.New("the expression factor must be wrapped in [], example: [A] > 0")
	}

	// 计算 TopN 或 BottomN 计算类型, 只支持单因子
	if data.CalculateType == 4 || data.CalculateType == 5 {
		if len(metricList) != 1 {
			return false, errors.New("the calculate_type 4 or 5 of TopN/BottomN rule types, it only supports single factor mode. example: [A] > 0")
		}
	}

	params := make(map[string]interface{})
	i := 0
	for _, metric := range metricList {
		params[metric[1]] = i
	}

	expr, err := govaluate.NewEvaluableExpression(data.Express)
	if err != nil {
		return false, errors.New("mathematical expression is incorrect, " + err.Error())
	}

	result, err := expr.Evaluate(params)
	if result == nil || err != nil {
		return false, errors.New("mathematical expression calculation is incorrect, " + err.Error())
	}

	// 对结果进行判定
	vaReflect := reflect.TypeOf(result)
	if strings.Compare(vaReflect.String(), "bool") != 0 {
		return false, errors.New("incorrect regular expression, example: [A] > 0")
	}

	// 关于时间窗口的校验
	for _, v := range data.TimeWindow {
		if len(v) != 2 {
			return false, errors.New("incorrect value written in time window")
		}

		for _, t := range v {
			_, err := time.ParseDuration(t)
			if err != nil {
				return false, errors.New("incorrect value written in time window")
			}
		}
	}

	return true, nil
}

// QueryRules 查询规则
func (r *mathRule) QueryRules(condition *apiModel.MathRuleCondition) (*[]apiModel.MathRule, int64, error) {
	var err error
	result := make([]apiModel.MathRule, 0)
	records, count, err := ruleDto.RuleDto.SelectByCondition(condition)
	if err == nil {
		for _, v := range *records {
			var window map[string][]string
			err := json.Unmarshal([]byte(v.TimeWindow), &window)
			if err == nil {
				// 告警接收人组
				var groupIds = make([]int, 0)
				for _, id := range strings.Split(v.GroupIp, ",") {
					groupIds = append(groupIds, util.StringToInt(id))
				}

				// 指标集
				var metrics = make(map[string]string)
				err := json.Unmarshal([]byte(v.MetricList), &metrics)
				if err == nil {
					result = append(result, apiModel.MathRule{
						Id:                 v.ID,
						Name:               v.Name,
						CalculateType:      v.CalculateType,
						Express:            v.Express,
						MetricList:         metrics,
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
						CreatedAt:          util.DateTimeToString(v.CreatedAt),
						UpdatedAt:          util.DateTimeToString(v.UpdatedAt),
					})
				}
			}
		}
	}

	return &result, count, err
}

// AddRule 添加规则
func (r *mathRule) AddRule(data *apiModel.MathRule) error {
	// 校验规则
	if _, err := r.CheckRule(data); err != nil {
		return err
	}

	window, _ := json.Marshal(data.TimeWindow)
	metrics, _ := json.Marshal(data.MetricList)
	var record = dbModel.Rule{
		Name:               data.Name,
		CalculateType:      data.CalculateType,
		Express:            data.Express,
		MetricList:         string(metrics),
		Threshold:          data.Threshold,
		TimeWindow:         string(window),
		Duration:           data.Duration,
		Origin:             data.Origin,
		BusinessType:       data.Type,
		Category:           data.Category,
		ExtensionCondition: data.ExtensionCondition,
		Level:              data.Level,
		Creator:            data.Creator,
		Updater:            data.Updater,
		ResponsiblePeople:  data.ResponsiblePeople,
		Crontab:            data.Crontab,
		Switch:             data.Switch,
		Inuse:              data.Inuse,
		GroupIp:            strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		WebHooks:           strings.Replace(strings.Trim(fmt.Sprint(data.WebHooks), "[]"), " ", ",", -1),
		Description:        data.Description,
		CreatedAt:          time.Now(),
	}

	if err := ruleDto.RuleDto.Insert(&record); err == nil {
		var ch = make(map[string]*apiModel.MathRule)
		ch["ADD"] = data
		calculate.MathSynchronizeRuleCh <- ch
		return nil
	} else {
		return err
	}
}

// UpdateRule 更新规则
func (r *mathRule) UpdateRule(data *apiModel.MathRule) error {
	// 新规则校验
	if _, err := r.CheckRule(data); err != nil {
		if !strings.Contains(err.Error(), "rule already exists") {
			return err
		}
	}

	window, _ := json.Marshal(data.TimeWindow)
	metrics, _ := json.Marshal(data.MetricList)
	var record = dbModel.Rule{
		ID:                 data.Id,
		Name:               data.Name,
		CalculateType:      data.CalculateType,
		Express:            data.Express,
		MetricList:         string(metrics),
		Threshold:          data.Threshold,
		Unit:               data.Unit,
		TimeWindow:         string(window),
		Duration:           data.Duration,
		Origin:             data.Origin,
		BusinessType:       data.Type,
		Category:           data.Category,
		ExtensionCondition: data.ExtensionCondition,
		Level:              data.Level,
		Creator:            data.Creator,
		Updater:            data.Updater,
		ResponsiblePeople:  data.ResponsiblePeople,
		Crontab:            data.Crontab,
		Switch:             data.Switch,
		Inuse:              data.Inuse,
		GroupIp:            strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		WebHooks:           strings.Replace(strings.Trim(fmt.Sprint(data.WebHooks), "[]"), " ", ",", -1),
		Description:        data.Description,
		UpdatedAt:          time.Now(),
	}

	err := ruleDto.RuleDto.Save(&record)
	if err == nil {
		var ch = make(map[string]*apiModel.MathRule)
		ch["UPDATE"] = data
		calculate.MathSynchronizeRuleCh <- ch
	}

	return err
}

// DeleteRule 删除规则
func (r *mathRule) DeleteRule(updater string, ids []int) error {
	// 查询记录, 获取需要删除的规则名称
	records, _, err := ruleDto.RuleDto.SelectByIds(ids)
	if err != nil {
		return err
	}

	err = ruleDto.RuleDto.Delete(updater, ids)
	if err == nil {
		for _, v := range *records {
			var data = apiModel.MathRule{
				Id:   v.ID,
				Name: v.Name,
			}

			var ch = make(map[string]*apiModel.MathRule)
			ch["DELETE"] = &data
			calculate.MathSynchronizeRuleCh <- ch
		}
	}

	return err
}

// DisableOrEnableForRule 禁用或开启规则
func (r *mathRule) DisableOrEnableForRule(id uint, status int8, updater string) (string, error) {
	if id == 0 {
		return "", errors.New("the rule id should be a positive integer")
	}

	if strings.Compare(updater, "") == 0 {
		return "", errors.New("the updater value of the rule must be specified")
	}

	// 1 --- 开启; 2 --- 禁用
	switch status {
	case 1:
		status = 1
	case 2:
		status = 2
	default:
		return "", errors.New("whether to enable, 1 --- on; 2 --- off")
	}

	// 先更新
	var data = dbModel.Rule{
		ID:      id,
		Switch:  status,
		Updater: updater,
	}
	var err error
	err = ruleDto.RuleDto.Update(&data)
	if err != nil {
		return "", err
	}

	var condition = apiModel.MathRuleCondition{
		Id: id,
	}
	records, count, err := ruleDto.RuleDto.SelectByCondition(&condition)
	if count > 0 && records != nil {
		for _, v := range *records {
			// 进行信号处理
			switch status {
			case 1: // 开启
				var metricList map[string]string
				if err = json.Unmarshal([]byte(v.MetricList), &metricList); err != nil {
					return "", err
				}

				var window map[string][]string
				if err = json.Unmarshal([]byte(v.TimeWindow), &window); err != nil {
					return "", err
				}

				var groupIds = make([]int, 0)
				for _, id := range strings.Split(v.GroupIp, ",") {
					groupIds = append(groupIds, util.StringToInt(id))
				}

				var ch = make(map[string]*apiModel.MathRule)
				ch["ADD"] = &apiModel.MathRule{
					Name:               v.Name,
					CalculateType:      v.CalculateType,
					Express:            v.Express,
					MetricList:         metricList,
					Threshold:          v.Threshold,
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
					Description:        data.Description,
					CreatedAt:          util.DateTimeToString(v.CreatedAt),
				}

				calculate.MathSynchronizeRuleCh <- ch
			case 2: // 禁用
				var ch = make(map[string]*apiModel.MathRule)
				ch["DELETE"] = &apiModel.MathRule{
					Id:   v.ID,
					Name: v.Name,
				}
				calculate.MathSynchronizeRuleCh <- ch
			}
		}
	}

	return "ok", err
}
