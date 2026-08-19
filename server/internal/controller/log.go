package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
)

// LogController 日志控制器（登录日志 + 操作日志）
type LogController struct{}

// @Summary 登录日志分页
// @Tags 日志
// @Security BearerAuth
// @Param query query dto.LoginLogListQuery true "-"
// @Success 200 {object} response.Result[dto.LoginLogListResp]
// @Router /log/login [get]
func (l *LogController) Login(c *gin.Context) {
	var query dto.LoginLogListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.LoginLogList(query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 操作日志分页
// @Tags 日志
// @Security BearerAuth
// @Param query query dto.OperationLogListQuery true "-"
// @Success 200 {object} response.Result[dto.OperationLogListResp]
// @Router /log/operation [get]
func (l *LogController) Operation(c *gin.Context) {
	var query dto.OperationLogListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.OperationLogList(query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 导出登录/操作日志（CSV）
// @Tags 日志
// @Security BearerAuth
// @Param type query string true "login | operation"
// @Success 200 {file} file
// @Router /log/export [get]
func (l *LogController) Export(c *gin.Context) {
	var query dto.LogExportQuery
	if !response.BindQuery(c, &query) {
		return
	}
	body, filename, err := service.LogExport(query.Type)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", body)
}
