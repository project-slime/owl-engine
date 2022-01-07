package router

const (
	srvGroupUri = "api/v0"

	// 系统健康路由
	ping   = "/health/ping"
	status = "/health/dashboard" // 系统运行资源消耗面板

	// 规则配置
	checkRule           = "/rule/checkRule"           // 规则校验
	queryRule           = "/rule/queryRule"           // 查询规则
	batchDeleteRule     = "/rule/batchDeleteRule"     // 批量删除规则
	deleteRule          = "/rule/deleteRule"          // 删除规则
	updateRule          = "/rule/updateRule"          // 更新规则
	addRule             = "/rule/addRule"             // 添加规则
	enableOrDisableRule = "/rule/enableOrDisableRule" // 禁用或开启规则

	checkLoggerRule           = "/rule/logger/checkRule"           // 规则校验
	queryLoggerRule           = "/rule/logger/queryRule"           // 查询规则
	batchDeleteLoggerRule     = "/rule/logger/batchDeleteRule"     // 批量删除规则
	deleteLoggerRule          = "/rule/logger/deleteRule"          // 删除规则
	updateLoggerRule          = "/rule/logger/updateRule"          // 更新规则
	addLoggerRule             = "/rule/logger/addRule"             // 添加规则
	enableOrDisableLoggerRule = "/rule/logger/enableOrDisableRule" // 禁用或开启规则
)
