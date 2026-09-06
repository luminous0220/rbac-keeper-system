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

// @Summary 导出登录日志（CSV，携带列表同款筛选）
// @Tags 日志
// @Security BearerAuth
// @Param query query LoginLogExportQuery false "-"
// @Success 200 {file} file
// @Router /log/login/export [get]
func (l *Controller) LoginExport(c *gin.Context) {
	var query LoginLogExportQuery
	if !response.BindQuery(c, &query) {
		return
	}
	body, filename, err := ExportLoginLogs(query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", body)
}

// @Summary 导出操作日志（CSV，携带列表同款筛选）
// @Tags 日志
// @Security BearerAuth
// @Param query query OperationLogExportQuery false "-"
// @Success 200 {file} file
// @Router /log/operation/export [get]
func (l *Controller) OperationExport(c *gin.Context) {
	var query OperationLogExportQuery
	if !response.BindQuery(c, &query) {
		return
	}
	body, filename, err := ExportOperationLogs(query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "text/csv; charset=utf-8", body)
}

// @Summary 清空登录日志（物理删除全部行）
// @Tags 日志
// @Security BearerAuth
// @Success 200 {object} response.Result[any]
// @Router /log/login/clear [post]
func (l *Controller) LoginClear(c *gin.Context) {
	n, err := ClearLoginLogs()
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	// 清空动作本身记一条操作日志（异步入队，便于审计谁清了库）
	CreateWithMq(c, "清空登录日志")
	response.OK(c, gin.H{"deleted": n})
}

// @Summary 清空操作日志（物理删除全部行）
// @Tags 日志
// @Security BearerAuth
// @Success 200 {object} response.Result[any]
// @Router /log/operation/clear [post]
func (l *Controller) OperationClear(c *gin.Context) {
	n, err := ClearOperationLogs()
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	CreateWithMq(c, "清空操作日志")
	response.OK(c, gin.H{"deleted": n})
}
