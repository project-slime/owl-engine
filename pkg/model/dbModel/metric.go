package dbModel

import "time"

// 指标值表
type Metric struct {
	Id         int                    `gorm:"column:id;type:int;AUTO_INCREMENT;PRIMARY_KEY"`
	MetricName string                 `gorm:"column:metric;type:varchar(128);NOT NULL"`  // 指标名
	Origin     string                 `gorm:"column:origin;type:varchar(128);NOT NULL"`  // 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	Type       string                 `gorm:"column:type;type:varchar(128);NOT NULL"`    // 类型,前端-异常、crash/业务-业务域/应用-异常、服务、JVM/组件-db、mq、redis/基础-网络、k8s、物理机、虚拟机
	Category   int                    `gorm:"column:category;type:int(2);NOT NULL"`      // 指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控
	Value      float32                `gorm:"column:value;type:double"`                  // 指标值
	Time       time.Time              `gorm:"column:time;type:timestamp;index"`          // 时间戳
	Creator    string                 `gorm:"column:creator;type:varchar(128);NOT NULL"` // 创建人,apiRobot/dbRobot/mqRobot
	Extension  map[string]interface{} `gorm:"column:extension;type:text"`                // 扩展字段
	CreatedAt  time.Time              `gorm:"column:created_at;type:timestamp;NOT NULL"` // 记录创建时间
}

// 指标值表
func (Metric) TableName() string {
	return "engine_tbl_metric"
}
