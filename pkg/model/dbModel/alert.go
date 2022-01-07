package dbModel

import (
	"time"
)

// 告警事件表
type Alert struct {
	ID           int       `gorm:"column:id;type:int;AUTO_INCREMENT;PRIMARY_KEY"`
	AlertId      string    `gorm:"column:alert_id;type:varchar(16);NOT NULL"`       // 告警事件的唯一id
	Name         string    `gorm:"column:name;type:varchar(255);NOT NULL"`          // 告警名称, 对应规则的名称
	Item         string    `gorm:"column:item;type:varchar(128);NOT NULL"`          // 告警项, 对应规则的表达式
	Origin       string    `gorm:"column:origin;type:varchar(128);NOT NULL;index"`  // 告警源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	BusinessType string    `gorm:"column:type;type:varchar(128);NOT NULL"`          // 告警子类型,前端-异常、crash/业务-业务域/应用-异常、服务、JVM/组件-db、mq、redis/基础-网络、k8s、物理机、虚拟机
	Category     int8      `gorm:"column:category;type:tinyint(1);NOT NULL"`        // 告警类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控
	Value        float64   `gorm:"column:value;type:double"`                        // 告警值
	Level        int8      `gorm:"column:level;type:tinyint(1);NOT NULL"`           // 告警级别:0-Not classified; 1-Information; 2-Warning; 3-critical; 4-Disaster
	Content      string    `gorm:"column:content;type:tinytext;NOT NULL"`           // 告警内容
	RuleName     string    `gorm:"column:rule_name;type:varchar(255);NOT NULL"`     // 规则名称
	GroupId      string    `gorm:"column:group_id;type:varchar(128);NOT NULL"`      // 告警联系组id, 多个id 以 , 进行分割
	Owner        string    `gorm:"column:owner;type:varchar(128)"`                  // 告警负责人
	Status       int8      `gorm:"column:status;type:tinyint(1)"`                   // 告警状态,1-告警中,2-恢复,3-忽略,4-静默
	Platform     int8      `gorm:"column:platform;type:tinyint(1)"`                 // 告警平台,1-owl,2-zcat,3-prometheus,4-zms等
	AlertTime    time.Time `gorm:"column:alert_time;type:timestamp;NOT NULL;index"` // 告警时间
	PlatformName string    `gorm:"column:platform_name;type:varchar;size:128"`      // 告警平台名称,zms/zdtp/es等
	AggregatorId int       `gorm:"column:aggregator_id;type:bigint"`                // 告警聚合id
	Creator      string    `gorm:"column:creator;type:varchar(64)"`                 // 创建人,engine/event
	Updater      string    `gorm:"column:updater;type:varchar(64)"`                 // 更改人
	CreatedAt    time.Time `gorm:"column:created_at;index"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Alert) TableName() string {
	return "engine_tbl_alert"
}
