package rule

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"owl-engine/pkg/xlogs"
	"strconv"
	"strings"
	"time"

	appConfig "owl-engine/pkg/config"
	"owl-engine/pkg/model/apiModel"
	"owl-engine/pkg/model/dbModel"
	"owl-engine/pkg/service/v0/calculate"
	"owl-engine/pkg/util"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/robfig/cron/v3"
)

/*
日志规则处理服务层。 因并发较小，因此并未充分考虑其高并发的安全性和资源消耗
*/
type loggerRule struct{}

var LoggerRuleSrv = new(loggerRule)

// CheckRule 规则合法性校验
func (l *loggerRule) CheckRule(data *apiModel.LoggerRule) (bool, error) {
	condition := apiModel.LoggerRuleCondition{
		Id:                data.Id,
		Name:              data.Name,
		Creator:           data.Creator,
		ResponsiblePeople: data.ResponsiblePeople,
		Switch:            data.Switch,
		Inuse:             data.Inuse,
	}

	rs := new(dbModel.LoggerRule)
	if _, count, _ := rs.Select(&condition); count > 0 {
		return false, errors.New("rule already exists for rule name " + data.Name)
	}

	// 关键字段值不能为空
	if strings.Compare(data.MessageField, "") == 0 {
		return false, errors.New("the message field cannot be empty")
	}

	// 关于 crontab 的表达式正则校验
	if _, err := cron.ParseStandard(data.Crontab); err != nil {
		return false, errors.New("cron express: " + err.Error())
	}

	// 关于es链接的验证
	if strings.Compare(data.Source, "es") != 0 {
		return false, errors.New("currently only supports es cluster to query data")
	}

	// 测试环境使用代理
	conf := appConfig.Get()
	esConfig := elasticsearch.Config{}
	if conf.ServerOptions.EnableProxy && strings.Compare(conf.ServerOptions.Proxy, "") != 0 {
		proxyUrl, _ := url.Parse(conf.ServerOptions.Proxy)
		esConfig = elasticsearch.Config{
			Addresses:  strings.Split(data.Address, ","),
			Username:   data.Username,
			Password:   data.Password,
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
		esConfig = elasticsearch.Config{
			Addresses:  strings.Split(data.Address, ","),
			Username:   data.Username,
			Password:   data.Password,
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
	es, err := elasticsearch.NewClient(esConfig)
	if err == nil {
		// 将 sql string 转换为 map[string]interface
		var query map[string]interface{}
		_ = json.Unmarshal([]byte(data.Sql), &query)

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err == nil {
			response, err := es.Search(es.Search.WithContext(context.Background()),
				es.Search.WithIndex(data.Index),
				es.Search.WithBody(&buf),
				es.Search.WithTrackTotalHits(true),
				es.Search.WithPretty())

			defer response.Body.Close()

			if err != nil || response.IsError() {
				return false, errors.New(fmt.Sprintf("query sql: %s error", data.Sql))
			}
		} else {
			return false, errors.New(fmt.Sprintf("query sql: %s error: %s", data.Sql, err.Error()))
		}
	} else {
		return false, errors.New(fmt.Sprintf("creating the elasticsearch client error: %s", err.Error()))
	}

	return true, nil
}

// QueryRule 查询规则
func (l *loggerRule) QueryRule(condition *apiModel.LoggerRuleCondition) (*[]apiModel.LoggerRule, int64, error) {
	result := make([]apiModel.LoggerRule, 0)

	loggerRule := new(dbModel.LoggerRule)
	records, count, err := loggerRule.Select(condition)
	if err != nil {
		xlogs.Errorf("query logger rules from table %s error, %s", loggerRule.TableName(), err.Error())
	} else {
		for _, value := range *records {
			ids := make([]int, 0)
			for _, idStr := range strings.Split(value.GroupIp, ",") {
				if id, err := strconv.Atoi(idStr); err == nil {
					ids = append(ids, id)
				}
			}

			result = append(result, apiModel.LoggerRule{
				Id:                value.ID,
				Name:              value.Name,
				Source:            value.Source,
				Address:           value.Address,
				Username:          value.Username,
				Password:          value.Password,
				Index:             value.Index,
				MessageField:      value.MessageField,
				Sql:               value.Sql,
				Threshold:         value.Threshold,
				Origin:            value.Origin,
				BusinessType:      value.BusinessType,
				Category:          value.Category,
				Level:             value.Level,
				Creator:           value.Creator,
				Updater:           value.Updater,
				ResponsiblePeople: value.ResponsiblePeople,
				Crontab:           value.Crontab,
				Switch:            value.Switch,
				Inuse:             value.Inuse,
				GroupId:           ids,
				Description:       value.Description,
				CreatedAt:         util.DateTimeToString(value.CreatedAt),
				UpdatedAt:         util.DateTimeToString(value.UpdatedAt),
			})
		}
	}

	return &result, count, nil
}

// AddRule 添加规则
func (l *loggerRule) AddRule(data *apiModel.LoggerRule) error {
	// 校验规则
	if _, err := l.CheckRule(data); err != nil {
		return err
	}

	rs := dbModel.LoggerRule{
		Name:              data.Name,
		Source:            data.Source,
		Address:           data.Address,
		Username:          data.Username,
		Password:          data.Password,
		Index:             data.Index,
		MessageField:      data.MessageField,
		Sql:               data.Sql,
		Threshold:         data.Threshold,
		Origin:            data.Origin,
		BusinessType:      data.BusinessType,
		Category:          data.Category,
		Level:             data.Level,
		Creator:           data.Creator,
		Updater:           data.Updater,
		ResponsiblePeople: data.ResponsiblePeople,
		Crontab:           data.Crontab,
		Switch:            data.Switch,
		Inuse:             data.Inuse,
		GroupIp:           strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		Description:       data.Description,
		CreatedAt:         time.Now(),
	}

	if err := rs.Insert(); err == nil {
		// 填充信号量
		ch := make(map[string]*apiModel.LoggerRule)
		ch["ADD"] = data
		calculate.LoggerRuleCh <- ch
		return nil
	} else {
		return err
	}
}

// UpdateRule 更新规则
func (l *loggerRule) UpdateRule(data *apiModel.LoggerRule) error {
	// 校验规则
	if _, err := l.CheckRule(data); err != nil {
		if !strings.Contains(err.Error(), "rule already exists") {
			return err
		}
	}

	rs := dbModel.LoggerRule{
		ID:                data.Id,
		Name:              data.Name,
		Source:            data.Source,
		Address:           data.Address,
		Username:          data.Username,
		Password:          data.Password,
		Index:             data.Index,
		MessageField:      data.MessageField,
		Sql:               data.Sql,
		Threshold:         data.Threshold,
		Origin:            data.Origin,
		BusinessType:      data.BusinessType,
		Category:          data.Category,
		Level:             data.Level,
		Creator:           data.Creator,
		Updater:           data.Updater,
		ResponsiblePeople: data.ResponsiblePeople,
		Crontab:           data.Crontab,
		Switch:            data.Switch,
		Inuse:             data.Inuse,
		GroupIp:           strings.Replace(strings.Trim(fmt.Sprint(data.GroupId), "[]"), " ", ",", -1),
		Description:       data.Description,
		UpdatedAt:         time.Now(),
	}

	if err := rs.Save(); err == nil {
		// 填充信号量
		ch := make(map[string]*apiModel.LoggerRule)
		ch["UPDATE"] = data
		calculate.LoggerRuleCh <- ch
		return nil
	} else {
		return err
	}
}

// DeleteRule 删除规则
func (l *loggerRule) DeleteRule(updater string, ids []int) error {
	rs := dbModel.LoggerRule{}

	// 查询记录, 获取需要删除的规则名称
	records, _, err := rs.SelectById(ids)
	if err != nil {
		return err
	}

	err = rs.Delete(updater, ids)
	if err == nil {
		for _, v := range *records {
			data := apiModel.LoggerRule{
				Id:   v.ID,
				Name: v.Name,
			}

			ch := make(map[string]*apiModel.LoggerRule)
			ch["DELETE"] = &data
			calculate.LoggerRuleCh <- ch
		}
	}

	return err
}

// DisableOrEnableForRule 禁用或开启规则
func (l *loggerRule) DisableOrEnableForRule(id uint, status int8, updater string) (string, error) {
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
	rs := dbModel.LoggerRule{
		ID:      id,
		Switch:  status,
		Updater: updater,
	}

	var err error
	err = rs.Update()
	if err != nil {
		return "", err
	}

	condition := apiModel.LoggerRuleCondition{
		Id: id,
	}

	records, count, err := rs.Select(&condition)
	if count > 0 && records != nil {
		for _, v := range *records {
			// 进行信号处理
			switch status {
			case 1: // 开启
				var groupIds = make([]int, 0)
				for _, id := range strings.Split(v.GroupIp, ",") {
					groupIds = append(groupIds, util.StringToInt(id))
				}

				ch := make(map[string]*apiModel.LoggerRule)
				ch["ADD"] = &apiModel.LoggerRule{
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
					GroupId:           groupIds,
					Description:       v.Description,
					CreatedAt:         util.DateTimeToString(v.CreatedAt),
				}

				calculate.LoggerRuleCh <- ch
			case 2: // 禁用
				ch := make(map[string]*apiModel.LoggerRule)
				ch["DELETE"] = &apiModel.LoggerRule{
					Id:   v.ID,
					Name: v.Name,
				}
				calculate.LoggerRuleCh <- ch
			}
		}
	}

	return "ok", err
}
