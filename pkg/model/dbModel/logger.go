package dbModel

import (
	database2 "owl-engine/pkg/client/database"
	"owl-engine/pkg/model/apiModel"
	"strings"
	"time"

	"gorm.io/gorm"
)

// LoggerRule 日志处理规则表
type LoggerRule struct {
	ID                uint           `gorm:"column:id;type:int;AUTO_INCREMENT;PRIMARY_KEY"`
	Name              string         `gorm:"column:name;type:varchar(255);NOT NULL;UNIQUE_INDEX"`  // 规则唯一名称
	Source            string         `gorm:"column:source;type:varchar(16);NOT NULL;default:es"`   // 日志的数据源, 默认为 elasticsearch
	Address           string         `gorm:"column:address;type:varchar(255)"`                     // 对于es等数据源，会需要连接地址，多个地址以 ',' 分隔
	Username          string         `gorm:"column:username;type:varchar(32)"`                     // 对于esc等数据源，其认证的用户名
	Password          string         `gorm:"column:password;type:varchar(32)"`                     // 对于es等数据源, 其需要认证的密码
	Index             string         `gorm:"index;type:varchar(128)"`                              // 对于 es 等数据源的索引, 支持模糊匹配
	MessageField      string         `gorm:"column:message_field;type:varchar(32);NOT NULL"`       // 告警时，需要查询到的告警内容的字段
	Sql               string         `gorm:"column:sql;type:json;NOT NULL"`                        // es 查询语句
	Threshold         float64        `gorm:"column:threshold;type:float;NOT NULL"`                 // 阈值
	Origin            string         `gorm:"column:origin;type:varchar(64);NOT NULL"`              // 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	BusinessType      string         `gorm:"column:business_type;type:varchar(64);NOT NULL"`       // 产品名: 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	Category          int8           `gorm:"column:category;type:tinyint(1);NOT NULL"`             // '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控'
	Level             int8           `gorm:"column:level;type:tinyint(1);default:3;NOT NULL"`      // 告警级别: 0 -- Not classified; 1 --- Information; 2 --- Warning; 3 --- critical; 4 --- Disaster
	Creator           string         `gorm:"column:creator;type:varchar(32);NOT NULL"`             // 规则创建者, 用户钉钉的 userid
	Updater           string         `gorm:"column:updater;type:varchar(32)"`                      // 规则创建者, 用户钉钉的 userid
	ResponsiblePeople string         `gorm:"column:responsible_people;type:varchar(255);NOT NULL"` // 告警时间的处理人, 用户钉钉的 userid
	Crontab           string         `gorm:"column:crontab;type:varchar(32);default:* * * * *"`    // 每条规则的定时任务执行表达式, 默认为: "* * * * *"
	Switch            int8           `gorm:"column:switch;type:tinyint(1);default:1"`              // 是否启用, 1 --- on; 2 --- off
	Inuse             int8           `gorm:"column:inuse;type:tinyint(1);default:2"`               // 是否删除, 1 --- yes; 2 --- no
	GroupIp           string         `gorm:"column:group_ip;type:varchar(255);NOT NULL"`           //  告警时间接收者的组id, 多个值以 ',' 分隔
	Description       string         `gorm:"column:description;type:tinytext(1024)"`               // 描述
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (LoggerRule) TableName() string {
	return "engine_tbl_logger_rules"
}

// 查询记录
func (l *LoggerRule) Select(condition *apiModel.LoggerRuleCondition) (*[]LoggerRule, int64, error) {
	db := database2.DB.Model(&LoggerRule{})

	if condition.Id > 0 {
		db = db.Where("id = ?", condition.Id)
	}

	if strings.Compare(condition.Name, "") != 0 {
		db = db.Where("name like ?", "%"+condition.Name+"%")
	}

	if strings.Compare(condition.Creator, "") != 0 {
		db = db.Where("creator = ?", condition.Creator)
	}

	if strings.Compare(condition.ResponsiblePeople, "") != 0 {
		db = db.Where("responsible_people = ?", condition.ResponsiblePeople)
	}

	if condition.Switch > 0 {
		db = db.Where("switch = ?", condition.Switch)
	}

	if condition.Inuse > 0 {
		db = db.Where("inuse = ?", condition.Inuse)
	}

	var count int64
	db.Count(&count)

	var record = make([]LoggerRule, 0, condition.Size)
	offset := (condition.Page - 1) * condition.Size

	// 按照更新时间进行排序
	return &record, count, db.Offset(int(offset)).Limit(int(condition.Size)).Order("updated_at desc").Scan(&record).Error
}

// 依据 ID 查询规则记录
func (l *LoggerRule) SelectById(ids []int) (*[]LoggerRule, int64, error) {
	db := database2.DB.Model(LoggerRule{}).Where("id in (?)", ids)

	var count int64
	db.Count(&count)

	var record = make([]LoggerRule, 0)
	return &record, count, db.Order("created_at desc").Scan(&record).Error
}

// 增加记录
func (l *LoggerRule) Insert() error {
	work := database2.NewWork()
	db := work.Begin()
	defer work.Rollback()

	err := db.Create(l).Error
	if err == nil {
		work.Commit()
	}

	return err
}

// 更新记录
func (l *LoggerRule) Update() error {
	work := database2.NewWork()
	db := work.Begin()
	defer work.Rollback()

	var err error
	var record LoggerRule
	err = db.Model(LoggerRule{}).Where("id = ?", l.ID).First(&record).Error
	if err == nil {
		record.Switch = l.Switch
		record.Updater = l.Updater

		err = db.Save(&record).Error
	}

	if err == nil {
		work.Commit()
	}

	return err
}

// gorm 使用 struct 类型对象作为参数时，
// struct会首先转化为map对象，然后再生成SQL语句，但是转化为map的过程中，对于零值字段会被忽略
// 致使其零值不会被更新
func (l *LoggerRule) Save() error {
	work := database2.NewWork()
	db := work.Begin()
	defer work.Rollback()

	var err error
	var record LoggerRule
	err = db.Model(LoggerRule{}).Where("id = ?", l.ID).First(&record).Error
	if err == nil {
		record.Name = l.Name
		record.Source = l.Source
		record.Address = l.Address
		record.Username = l.Username
		record.Password = l.Password
		record.Index = l.Index
		record.MessageField = l.MessageField
		record.Sql = l.Sql
		record.Threshold = l.Threshold
		record.Origin = l.Origin
		record.BusinessType = l.BusinessType
		record.Category = l.Category
		record.Level = l.Level
		record.Creator = l.Creator
		record.Updater = l.Updater
		record.ResponsiblePeople = l.ResponsiblePeople
		record.Crontab = l.Crontab
		record.Switch = l.Switch
		record.Inuse = l.Inuse
		record.GroupIp = l.GroupIp
		record.Description = l.Description
		record.UpdatedAt = l.UpdatedAt

		err = db.Save(&record).Error
	}

	if err == nil {
		work.Commit()
	}

	return err
}

// 删除
func (l *LoggerRule) Delete(updater string, ids []int) error {
	work := database2.NewWork()
	db := work.Begin()
	defer work.Rollback()

	var err error
	err = db.Model(LoggerRule{}).Where("id in (?)", ids).UpdateColumn("updater", updater).Error
	err = db.Model(LoggerRule{}).Where("id in (?)", ids).UpdateColumn("deleted_at", time.Now()).Error

	if err == nil {
		work.Commit()
	}

	return err
}
