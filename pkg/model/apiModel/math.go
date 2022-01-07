package apiModel

// MathRule 数学规则接口响应参数
type MathRule struct {
	Id                 uint                `json:"id"`
	Name               string              `json:"name"`
	CalculateType      int                 `json:"calculate_type"`      // 计算类型: 1 -- 最大值; 2 -- 最小值; 3 -- 环比; 4 -- TopN; 5 -- BottomN
	Express            string              `json:"express"`             // 计算表达式
	MetricList         map[string]string   `json:"metric_list"`         // 指标名集
	Threshold          float64             `json:"threshold"`           // 阈值, 可为零值
	Unit               string              `json:"unit"`                // 单位
	TimeWindow         map[string][]string `json:"time_window"`         // 时间窗口
	Duration           int                 `json:"duration"`            // 持续次数
	Origin             string              `json:"origin"`              // 产品名: '来源，前端-产品/业务-产品/应用-appid/组件-ip、集群名/基础-域名、ip'
	Type               string              `json:"type"`                // 业务域: '类型,前端-异常、crash/业务-业务域/应用-异常、服务、JVM/组件-db、mq、redis/基础-网络、k8s、物理机、虚拟机'
	Category           int8                `json:"category"`            // '指标类型,1-前端监控,2-业务监控,3-应用监控,4-组件监控,5-基础监控'
	ExtensionCondition string              `json:"extension_condition"` // 扩展条件
	Level              int8                `json:"level"`               // 告警级别: 0 -- Not classified; 1 --- Information; 2 --- Warning; 3 --- critical; 4 --- Disaster
	Creator            string              `json:"creator"`             // 规则创建者, 用户钉钉的 userid
	Updater            string              `json:"updater"`             // 规则的更新者,用户钉钉的 userid
	ResponsiblePeople  string              `json:"responsible_people"`  // 告警时间的处理人, 用户钉钉的 userid
	Crontab            string              `json:"crontab"`             // 每条规则的定时任务执行表达式, 默认为: "* * * * *"
	Switch             int8                `json:"switch"`              // 是否启用, 1 --- on; 2 --- off
	Inuse              int8                `json:"inuse"`               // 是否删除, 1 --- yes; 2 --- no
	GroupId            []int               `json:"group_id"`            //  告警时间接收者的组id
	WebHooks           []string            `json:"web_hooks"`
	Description        string              `json:"description"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
}

// MathRuleCondition 数学规则查询条件接口参数
type MathRuleCondition struct {
	Id                uint   `form:"id"`
	Name              string `form:"name"`
	Creator           string `form:"creator"`
	ResponsiblePeople string `form:"responsible_people"` // 告警时间的处理人, 用户钉钉的 userid
	Origin            string `form:"origin"`
	Type              string `form:"type"`
	Category          int8   `form:"category"`
	Switch            int8   `form:"switch"` // 是否启用, 1 --- on; 2 --- off
	Inuse             int8   `json:"inuse"`  // 是否删除, 1 --- yes; 2 --- no
	Page              int64  `form:"page" binding:"required,page_and_size"`
	Size              int64  `form:"size" binding:"required,page_and_size"`
}
