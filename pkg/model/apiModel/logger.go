package apiModel

// 日志规则接口响应参数
type LoggerRule struct {
	Id                uint    `json:"id"`
	Name              string  `json:"name"`
	Source            string  `json:"source"`             // 数据源, 当前默认从 elasticsearch
	Address           string  `json:"address"`            // elasticsearch 的连接地址, 多个地址, 以 ',' 分隔
	Username          string  `json:"username"`           // elasticsearch 的用户名
	Password          string  `json:"password"`           // elasticsearch 的密码
	Index             string  `json:"index"`              // elasticsearch 的索引, 支持模糊匹配
	MessageField      string  `json:"message_field"`      // elasticsearch 中的告警记录的字段
	Sql               string  `json:"sql"`                // Es 的查询语句
	Threshold         float64 `json:"threshold"`          // 阈值
	Origin            string  `json:"origin"`             // 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	BusinessType      string  `json:"business_type"`      // 产品名: 来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip
	Category          int8    `json:"category"`           // '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控'
	Level             int8    `json:"level"`              // 告警级别: 0 -- Not classified; 1 --- Information; 2 --- Warning; 3 --- critical; 4 --- Disaster
	Creator           string  `json:"creator"`            // 规则创建者, 用户钉钉的 userid
	Updater           string  `json:"updater"`            // 规则的更新者,用户钉钉的 userid
	ResponsiblePeople string  `json:"responsible_people"` // 告警时间的处理人, 用户钉钉的 userid
	Crontab           string  `json:"crontab"`            // 每条规则的定时任务执行表达式, 默认为: "* * * * *"
	Switch            int8    `json:"switch"`             // 是否启用, 1 --- on; 2 --- off
	Inuse             int8    `json:"inuse"`              // 是否删除, 1 --- yes; 2 --- no
	GroupId           []int   `json:"group_id"`           //  告警时间接收者的组id
	Description       string  `json:"description"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// 日志规则查询条件接口参数
type LoggerRuleCondition struct {
	Id                uint   `form:"id"`
	Name              string `form:"name"`
	Creator           string `form:"creator"`
	ResponsiblePeople string `form:"responsible_people"` // 告警时间的处理人, 用户钉钉的 userid
	Switch            int8   `form:"switch"`             // 是否启用, 1 --- on; 2 --- off
	Inuse             int8   `json:"inuse"`              // 是否删除, 1 --- yes; 2 --- no
	Page              int64  `form:"page" binding:"required,page_and_size"`
	Size              int64  `form:"size" binding:"required,page_and_size"`
}
