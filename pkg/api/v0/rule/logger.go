package rule

import (
	"strings"

	"owl-engine/pkg/model/apiModel"
	ruleSrv "owl-engine/pkg/service/v0/rule"
	"owl-engine/pkg/util"
	"owl-engine/pkg/util/resp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type loggerRule struct{}

var LoggerRule = new(loggerRule)

// CheckRule 规则校验
func (l *loggerRule) CheckRule(ctx *gin.Context) {
	var rs = new(apiModel.LoggerRule)

	if err := ctx.ShouldBindJSON(rs); err == nil {
		if _, err := ruleSrv.LoggerRuleSrv.CheckRule(rs); err == nil {
			resp.SuccessResp(ctx, "0", "logger rule check success")
		} else {
			resp.ErrorResp(ctx, "1", err.Error())
		}
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}

	return
}

// AddRule 规则添加
func (l *loggerRule) AddRule(ctx *gin.Context) {
	var rs = new(apiModel.LoggerRule)

	if err := ctx.ShouldBindJSON(rs); err == nil {
		if err := ruleSrv.LoggerRuleSrv.AddRule(rs); err == nil {
			resp.SuccessResp(ctx, "0", "add logger rule ok")
		} else {
			resp.ErrorResp(ctx, "1", err.Error())
		}
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}

	return
}

// QueryRule 查询规则
func (l *loggerRule) QueryRule(ctx *gin.Context) {
	var condition = new(apiModel.LoggerRuleCondition)
	var result = struct {
		Page  int64                 `json:"page"`
		Size  int64                 `json:"size"`
		Total int64                 `json:"total"`
		Data  []apiModel.LoggerRule `json:"data"`
	}{
		Data: make([]apiModel.LoggerRule, 0),
	}

	var err error
	err = ctx.ShouldBindWith(condition, binding.Query)
	if err == nil {
		data, count, err := ruleSrv.LoggerRuleSrv.QueryRule(condition)
		if err == nil {
			result.Page = condition.Page
			result.Size = condition.Size
			result.Total = count
			result.Data = *data
		}
	}

	if err == nil {
		resp.SuccessJsonResp(ctx, "0", "query rules success", result)
	} else {
		resp.SuccessJsonResp(ctx, "1", err.Error(), result)
	}

	return
}

// UpdateRule 更新规则
func (l *loggerRule) UpdateRule(ctx *gin.Context) {
	var rs = new(apiModel.LoggerRule)

	if err := ctx.ShouldBindJSON(rs); err == nil {
		if err := ruleSrv.LoggerRuleSrv.UpdateRule(rs); err == nil {
			resp.SuccessResp(ctx, "0", "the logger update success")
		} else {
			resp.ErrorResp(ctx, "1", "the logger rule update error: "+err.Error())
		}
	} else {
		resp.ErrorResp(ctx, "1", "the param deserialization json error: "+err.Error())
	}

	return
}

// DeleteRule 单条删除规则
func (l *loggerRule) DeleteRule(ctx *gin.Context) {
	idStr := ctx.Request.FormValue("id")
	if strings.Compare(idStr, "") == 0 {
		resp.ErrorResp(ctx, "1", "the id value must be specified")
		ctx.Abort()
		return
	}

	id := util.StringToInt(idStr)

	updater := ctx.Request.FormValue("updater")
	if strings.Compare(updater, "") == 0 {
		resp.ErrorResp(ctx, "1", "the updater value must be specified")
		ctx.Abort()
		return
	}

	var ids = make([]int, 0)
	ids = append(ids, id)
	if err := ruleSrv.LoggerRuleSrv.DeleteRule(updater, ids); err == nil {
		resp.SuccessResp(ctx, "0", "ok")
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// BatchDeleteRule 批量删除
func (l *loggerRule) BatchDeleteRule(ctx *gin.Context) {
	idStr := ctx.QueryArray("id")

	var ids = make([]int, 0)
	for _, id := range idStr {
		ids = append(ids, util.StringToInt(id))
	}

	updater := ctx.Query("updater")
	if strings.Compare(updater, "") == 0 {
		resp.ErrorResp(ctx, "1", "the updater value must be specified")
		ctx.Abort()
		return
	}

	if err := ruleSrv.LoggerRuleSrv.DeleteRule(updater, ids); err == nil {
		resp.SuccessResp(ctx, "0", "ok")
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// EnableOrDisableRule 开启或禁用规则
func (l *loggerRule) EnableOrDisableRule(ctx *gin.Context) {
	idStr, ok := ctx.GetQuery("id")
	if !ok || strings.Compare(idStr, "") == 0 {
		resp.ErrorResp(ctx, "1", "the id of the rule must be specified")
		ctx.Abort()
		return
	}
	id := util.StringToInt(idStr)

	statusStr, ok := ctx.GetQuery("switch")
	if !ok || strings.Compare(statusStr, "") == 0 {
		resp.ErrorResp(ctx, "1", "the switch value of the rule must be specified")
		ctx.Abort()
		return
	}

	status := util.StringToInt(statusStr)

	updater, ok := ctx.GetQuery("updater")
	if !ok || strings.Compare(updater, "") == 0 {
		resp.ErrorResp(ctx, "1", "the updater value of the rule must be specified")
		ctx.Abort()
		return
	}

	if msg, err := ruleSrv.LoggerRuleSrv.DisableOrEnableForRule(uint(id), int8(status), updater); err == nil {
		resp.SuccessResp(ctx, "0", msg)
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}

	return
}
