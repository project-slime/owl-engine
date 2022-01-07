package constParam

// http header constant
const (
	ContentTypeJson  = "application/json"
	ContentAcceptAll = "*/*"
)

// common constant
const (
	NULL  = "" // 空字符串
	INUSE = 1  // 可用
	UNUSE = 0  // 不可用
)

// time format constant
const (
	DateMonthFormat        = "2006-01"
	DateFormat             = "2006-01-02"
	DateTimeMinFormat      = "2006-01-02 15:04"
	DateTimeFormat         = "2006-01-02 15:04:05"
	DateTimeWholeMinFormat = "2006-01-02 15:04:00"
)

// symbol constant
const (
	SymbolSemicolon    = ";"
	SymbolDash         = "-"
	SymbolComma        = ","
	SymbolBracketLeft  = "【"
	SymbolBracketRight = "】"
	SymbolSpace        = " "
	SymbolNewLine      = "\n"
	SymbolQueryNull    = "N/A"
)

// response constant
const (
	StatusSuccess = true
	StatusFail    = false

	StatusCodeOk        = "0000"
	StatusCodeParamErr  = "0001"
	StatusCodeCreateErr = "0002"
	StatusCodeUpdateErr = "0003"
	StatusCodeQueryErr  = "0004"
	StatusCodeDeleteErr = "0005"
	StatusCodeTypeErr   = "0006"

	RespOk        = "请求成功"
	RespParamErr  = "参数错误"
	RespCreateErr = "新增失败"
	RespUpdateErr = "更新失败"
	RespQueryErr  = "查询失败"
	RespDeleteErr = "删除失败"
	RespTypeErr   = "非法请求"
)

// api constant
const (
	// 指标
	PutMetric     = "指标传入"
	PutMetricList = "指标集传入"
)

// type constant
const (
	ApiCreator = "apiRobot"
	DbCreator  = "dbRobot"
	MqCreator  = "mqRobot"
)

// switch constant
const (
	SwitchOn  = 1
	SwitchOff = 0
)

// rule constant
const (
	AlertBizTemplateDingTalk   = "[%s][夜枭监控平台] \\n 告警名称：%s \\n 告警类型：业务告警 \\n 业务域：%s \\n 告警源：%s \\n 告警内容：%v分钟内，%s \\n 告警值：%v \\n 告警时间：%v \\n 负责人：%s"
	AlertBizTemplateWorkNotice = "## [%s][夜枭监控平台] \\n #### 告警名称：%s \\n #### 告警类型：业务告警 \\n #### 业务域：%s \\n #### 告警源：%s \\n #### 告警内容：%v分钟内，%s \\n #### 告警值：%v \\n #### 告警时间：%v \\n #### 负责人：%s"
	AlertBizTemplateTitle      = "[%s][夜枭监控平台]"
)

// alert constant
const (
	AlertLevelDisaster = 1
	AlertLevelCritical = 2
	AlertLevelWarning  = 3
	AlertLevelInfo     = 4

	AlertLevelDescDisaster  = "Disaster"
	AlertLevelDescCritical  = "Critical"
	AlertLevelDescWarning   = "Warning"
	AlertLevelDescInfo      = "Info"
	AlertLevelDescUndefined = "TypeUndefined"

	AlertPlatformZCAT = 1
	AlertPlatformProm = 2
	AlertPlatformOwl  = 3
	AlertPlatformZMS  = 4

	AlertStatusAlarming = 1
	AlertStatusRecover  = 2
	AlertStatusIgnore   = 3

	AlertCreatorOwl = "owl-engine"

	AlertFlagSuccess = "1"
	AlertFlagFail    = "0"

	AlertMetricMapFinance = `{"income":"收入","outcome":"支出","ztBalance":"中天余额"}`

	AlertReceiverTemplate = "%s(%s|%s)"
)
