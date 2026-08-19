package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rbackeeper/initialize"
	"rbackeeper/response"
)

// HealthController 健康检查控制器
type HealthController struct{}

// @Summary 存活探针(进程存活即返回 200)
// @Tags 健康
// @Success 200 {object} response.Result[any]
// @Router /healthz [get]
func (h *HealthController) Healthz(ctx *gin.Context) {
	response.OK(ctx, gin.H{"status": "ok"})
}

// @Summary 就绪探针(探活所有依赖组件，全部健康返回 200，否则 503)
// @Tags 健康
// @Success 200 {object} response.Result[any]
// @Router /readyz [get]
func (h *HealthController) Readyz(ctx *gin.Context) {
	comps, ok := initialize.HealthCheck()
	if ok {
		response.OK(ctx, comps)
		return
	}
	ctx.JSON(http.StatusServiceUnavailable, response.Result[any]{
		Code: response.ErrServer,
		Msg:  "部分依赖异常",
		Data: comps,
	})
}
