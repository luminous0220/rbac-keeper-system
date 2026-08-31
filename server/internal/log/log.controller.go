package log

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/response"
)

// Controller 日志控制器（登录日志 + 操作日志）
type Controller struct{}

// @Summary 登录日志分页
// @Tags 日志
// @Security BearerAuth
// @Param query query LoginLogListQuery true "-"
// @Success 200 {object} response.Result[LoginLogListResp]
// @Router /log/login [get]
func (l *Controller) Login(c *gin.Context) {
	var query LoginLogListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := LoginLogList(query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 操作日志分页
// @Tags 日志
// @Security BearerAuth
// @Param query query OperationLogListQuery true "-"
// @Success 200 {object} response.Result[OperationLogListResp]
// @Router /log/operation [get]
func (l *Controller) Operation(c *gin.Context) {
	var query OperationLogListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := OperationLogList(query)
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
func (l *Controller) Export(c *gin.Context) {
	var query LogExportQuery
	if !response.BindQuery(c, &query) {
		return
	}
	body, filename, err := Export(query.Type)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", body)
}
