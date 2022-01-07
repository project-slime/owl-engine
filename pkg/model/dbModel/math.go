package dbModel

import (
	"time"

	"gorm.io/gorm"
)

// Rule 数学运算规则表
type Rule struct {
	ID                 uint           `gorm:"column:id;type:int;AUTO_INCREMENT;PRIMARY_KEY"`
	Name               string         `gorm:"column:name;type:varchar(255);NOT NULL;UNIQUE_INDEX"`  // 规则唯一名称
	CalculateType      int            `gorm:"column:calculate_type;type:tinyint(1);NOT NULL"`       // 计算类型: 1 -- 最大值; 2 -- 最小值; 3 -- 环比; 4 -- TopN; 5 -- BottomN
	Express            string         `gorm:"column:express;type:tinytext(512);NOT NULL"`           // 计算表达式
	MetricList         string         `gorm:"column:metric_list;type:tinytext;NOT NULL"`            // 指标名集合
	Threshold          float64        `gorm:"column:threshold;type:float;default:0.0"`              // 阈值, 可为零值
	Unit               string         `gorm:"column:unit;type:varchar(16)"`                         // 单位
	TimeWindow         string         `gorm:"column:time_window;type:varchar(255)"`                 // 时间窗口, 默认都以 分钟 作为单位
	Duration           int            `gorm:"column:duration;type:tinyint(1);default:1"`            // 持续的次数在
	Origin             string         `gorm:"column:origin;type:varchar(64);NOT NULL"`              // 产品名: '来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip'
	BusinessType       string         `gorm:"column:business_type;type:varchar(64);NOT NULL"`       // 业务域: '类型,前端-异常、crash/业务-业务域/应用-异常、服务、JVM/组件-db、mq、redis/基础-网络、k8s、物理机、虚拟机'
	Category           int8           `gorm:"column:category;type:tinyint(1);NOT NULL"`             // '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控'
	ExtensionCondition string         `gorm:"column:extension_condition;type:varchar(255)"`         // 扩展条件
	Level              int8           `gorm:"column:level;type:tinyint(1);NOT NULL"`                // 告警级别: 0 -- Not classified; 1 --- Information; 2 --- Warning; 3 --- critical; 4 --- Disaster
	Creator            string         `gorm:"column:creator;type:varchar(32);NOT NULL"`             // 规则创建者, 用户钉钉的 userid
	Updater            string         `gorm:"column:updater;type:varchar(32)"`                      // 规则创建者, 用户钉钉的 userid
	ResponsiblePeople  string         `gorm:"column:responsible_people;type:varchar(255);NOT NULL"` // 告警时间的处理人, 用户钉钉的 userid
	Crontab            string         `gorm:"column:crontab;type:varchar(32);default:* * * * *"`    // 每条规则的定时任务执行表达式, 默认为: "* * * * *"
	Switch             int8           `gorm:"column:switch;type:tinyint(1);default:1"`              // 是否启用, 1 --- on; 2 --- off
	Inuse              int8           `gorm:"column:inuse;type:tinyint(1);default:2"`               // 是否删除, 1 --- yes; 2 --- no
	GroupIp            string         `gorm:"column:group_ip;type:varchar(255);NOT NULL"`           //  告警时间接收者的组id, 多个值以 ',' 分隔
	WebHooks           string         `gorm:"column:web_hooks;type:tinytext(1024)"`                 // 告警的 hook 地址,  多个值以 ',' 分隔
	Description        string         `gorm:"column:description;type:tinytext(1024)"`               // 描述
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Rule) TableName() string {
	return "engine_tbl_rules"
}
