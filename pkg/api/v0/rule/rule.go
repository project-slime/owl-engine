package rule

import (
	"owl-engine/pkg/model/apiModel"
	ruleSrv "owl-engine/pkg/service/v0/rule"
	"owl-engine/pkg/util"
	"owl-engine/pkg/util/resp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type rule struct{}

var Rule = new(rule)

// CheckRule 规则校验
func (r *rule) CheckRule(ctx *gin.Context) {
	var rs apiModel.MathRule

	if err := ctx.ShouldBindJSON(&rs); err == nil {
		if ok, err := ruleSrv.MathRuleSrv.CheckRule(&rs); !ok {
			resp.ErrorResp(ctx, "1", err.Error())
		} else {
			resp.SuccessResp(ctx, "0", "ok")
		}
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// AddRule 规则添加
func (r *rule) AddRule(ctx *gin.Context) {
	var rs apiModel.MathRule

	if err := ctx.ShouldBindJSON(&rs); err == nil {
		if err := ruleSrv.MathRuleSrv.AddRule(&rs); err == nil {
			resp.SuccessResp(ctx, "0", "ok")
		} else {
			resp.ErrorResp(ctx, "1", err.Error())
		}
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// QueryRule 查询规则
func (r *rule) QueryRule(ctx *gin.Context) {
	var condition apiModel.MathRuleCondition

	var result = struct {
		Page  int64               `json:"page"`
		Size  int64               `json:"size"`
		Total int64               `json:"total"`
		Data  []apiModel.MathRule `json:"data"`
	}{
		Data: make([]apiModel.MathRule, 0),
	}
	var err error
	err = ctx.ShouldBindWith(&condition, binding.Query)
	if err == nil {
		record, count, err := ruleSrv.MathRuleSrv.QueryRules(&condition)
		if err == nil {
			result.Page = condition.Page
			result.Size = condition.Size
			result.Total = count
			result.Data = *record

			resp.SuccessJsonResp(ctx, "0", "ok", result)
			return
		}
	}

	resp.ErrorResp(ctx, "1", err.Error())
}

// UpdateRule 更新规则
func (r *rule) UpdateRule(ctx *gin.Context) {
	var rs apiModel.MathRule

	if err := ctx.ShouldBindJSON(&rs); err == nil {
		if err := ruleSrv.MathRuleSrv.UpdateRule(&rs); err == nil {
			resp.SuccessResp(ctx, "0", "ok")
		} else {
			resp.ErrorResp(ctx, "1", err.Error())
		}
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// DeleteRule 单条删除规则
func (r *rule) DeleteRule(ctx *gin.Context) {
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
	if err := ruleSrv.MathRuleSrv.DeleteRule(updater, ids); err == nil {
		resp.SuccessResp(ctx, "0", "ok")
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// BatchDeleteRule 批量删除
func (r *rule) BatchDeleteRule(ctx *gin.Context) {
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

	if err := ruleSrv.MathRuleSrv.DeleteRule(updater, ids); err == nil {
		resp.SuccessResp(ctx, "0", "ok")
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}
}

// EnableOrDisableRule 开启或禁用规则
func (r *rule) EnableOrDisableRule(ctx *gin.Context) {
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

	if msg, err := ruleSrv.MathRuleSrv.DisableOrEnableForRule(uint(id), int8(status), updater); err == nil {
		resp.SuccessResp(ctx, "0", msg)
	} else {
		resp.ErrorResp(ctx, "1", err.Error())
	}

	return
}
