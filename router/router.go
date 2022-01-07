package router

import (
	"fmt"
	"owl-engine/pkg/xlogs"

	"owl-engine/pkg/api/common"
	"owl-engine/pkg/api/v0/healthy"

	"owl-engine/pkg/api/v0/rule"
	"owl-engine/router/middleware"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// loca 通常取决于 http 请求头的 'Accept-Language'
func translatorInit(local string) (err error) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New() //chinese
		enT := en.New() //english
		uni := ut.New(enT, zhT, enT)

		var success bool
		trans, success := uni.GetTranslator(local)
		if !success {
			return fmt.Errorf("uni.GetTranslator(%s) failed", local)
		}

		err = v.RegisterValidation("page_and_size", common.PageAndSizeValid)

		// 注册翻译器
		switch local {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
	}

	return
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	// pprof 性能分析
	pprof.Register(router)

	// 中间件的使用
	middleware.InitMiddleware(router)

	//router.MaxMultipartMemory = 8 << 20

	// 自定义校验器
	if err := translatorInit("zh"); err != nil {
		xlogs.Errorf("client trans failed, err:%v\n", err)
		return nil
	}

	// 系统运行状态
	healthGroup := router.Group("")
	{
		healthGroup.GET(ping, healthy.Pong)
		healthGroup.GET(status, healthy.Status)
	}

	// 规则
	ruleGroup := router.Group(srvGroupUri).Use(middleware.Auth())
	{
		ruleGroup.GET(checkRule, rule.Rule.CheckRule)
		ruleGroup.POST(addRule, rule.Rule.AddRule)
		ruleGroup.GET(queryRule, rule.Rule.QueryRule)
		ruleGroup.POST(updateRule, rule.Rule.UpdateRule)
		ruleGroup.DELETE(deleteRule, rule.Rule.DeleteRule)
		ruleGroup.DELETE(batchDeleteRule, rule.Rule.BatchDeleteRule)
		ruleGroup.POST(enableOrDisableRule, rule.Rule.EnableOrDisableRule)
	}

	// 日志处理规则
	logGroup := router.Group(srvGroupUri).Use(middleware.Auth())
	{
		logGroup.POST(checkLoggerRule, rule.LoggerRule.CheckRule)
		logGroup.POST(addLoggerRule, rule.LoggerRule.AddRule)
		logGroup.GET(queryLoggerRule, rule.LoggerRule.QueryRule)
		logGroup.POST(updateLoggerRule, rule.LoggerRule.UpdateRule)
		logGroup.DELETE(deleteLoggerRule, rule.LoggerRule.DeleteRule)
		logGroup.DELETE(batchDeleteLoggerRule, rule.LoggerRule.BatchDeleteRule)
		logGroup.POST(enableOrDisableLoggerRule, rule.LoggerRule.EnableOrDisableRule)
	}

	return router
}
