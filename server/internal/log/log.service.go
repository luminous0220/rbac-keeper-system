package log

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"

	"rbackeeper/global"
	"rbackeeper/model"
	"rbackeeper/utils"
)

// LoginLogList 登录日志分页
func LoginLogList(query LoginLogListQuery) (*LoginLogListResp, error) {
	rows, total, err := GetLoginLogList(query)
	if err != nil {
		return nil, err
	}
	list := make([]LoginLogItem, 0, len(rows))
	for i := range rows {
		list = append(list, toLoginLogItem(&rows[i]))
	}
	return &LoginLogListResp{
		Total:    total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize,
		List:     list,
	}, nil
}

// OperationLogList 操作日志分页
func OperationLogList(query OperationLogListQuery) (*OperationLogListResp, error) {
	rows, total, err := GetOperationLogList(query)
	if err != nil {
		return nil, err
	}
	list := make([]OperationLogItem, 0, len(rows))
	for i := range rows {
		list = append(list, toOperationLogItem(&rows[i]))
	}
	return &OperationLogListResp{
		Total:    total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize,
		List:     list,
	}, nil
}

// ExportLoginLogs 按筛选导出登录日志 CSV；返回字节流与建议文件名。
// 行数上限由 dao.logExportLimit 约束（防 OOM）。
func ExportLoginLogs(q LoginLogExportQuery) ([]byte, string, error) {
	rows, err := LoginLogExportRows(q)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "用户名", "IP", "归属", "浏览器", "系统", "状态", "消息", "登录时间"})
	for i := range rows {
		r := &rows[i]
		_ = w.Write([]string{
			fmt.Sprintf("%d", r.ID), r.Username, r.IP, r.IPAddr, r.Browser, r.OS,
			statusText(r.Status), r.Msg, r.LoginTime.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	return buf.Bytes(), "login_logs.csv", nil
}

// ExportOperationLogs 按筛选导出操作日志 CSV；返回字节流与建议文件名。
func ExportOperationLogs(q OperationLogExportQuery) ([]byte, string, error) {
	rows, err := OperationLogExportRows(q)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "模块", "方法", "操作人", "URL", "IP", "状态", "错误", "操作时间"})
	for i := range rows {
		r := &rows[i]
		_ = w.Write([]string{
			fmt.Sprintf("%d", r.ID), r.Title, r.RequestMethod,
			r.OperName, r.OperURL, r.OperIP,
			statusText(r.Status), r.ErrorMsg, r.OperTime.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	return buf.Bytes(), "operation_logs.csv", nil
}

func toLoginLogItem(r *model.LoginLog) LoginLogItem {
	return LoginLogItem{
		ID:        r.ID,
		Username:  r.Username,
		IP:        r.IP,
		IPAddr:    r.IPAddr,
		Browser:   r.Browser,
		OS:        r.OS,
		Status:    r.Status,
		Msg:       r.Msg,
		LoginTime: r.LoginTime,
	}
}

func toOperationLogItem(r *model.OperationLog) OperationLogItem {
	return OperationLogItem{
		ID:            r.ID,
		Title:         r.Title,
		RequestMethod: r.RequestMethod,
		OperName:      r.OperName,
		OperURL:       r.OperURL,
		OperIP:        r.OperIP,
		Status:        r.Status,
		ErrorMsg:      r.ErrorMsg,
		OperTime:      r.OperTime,
	}
}

func statusText(s int8) string {
	if s == 0 {
		return "成功"
	}
	return "失败"
}

// CreateWithMq 异步入队一条“成功”操作日志（status=0）。
// 用法：controller 业务成功 return 前调一次。
func CreateWithMq(c *gin.Context, title string) {
	CreateWithMqResult(c, title, 0, "")
}

// CreateWithMqResult 异步入队一条操作日志，可指定状态与错误消息。
// 失败仅 Warn，不阻断业务。越权 403 时以 status=1 调用，记录异常态。
func CreateWithMqResult(c *gin.Context, title string, status int8, errorMsg string) {
	// 生产者未初始化（如未配置 RocketMQ 直接启动）时静默跳过，不阻断业务
	if global.RocketMQ == nil || global.CFG == nil {
		return
	}
	claims := utils.GetClaims(c)
	operName := ""
	if claims != nil {
		operName = claims.Username
	}
	p := OperationLogPayload{
		Title:         title,
		RequestMethod: c.Request.Method,
		OperName:      operName,
		OperURL:       c.Request.URL.Path,
		OperIP:        c.ClientIP(),
		OperParam:     maskedRequestParam(c),
		Status:        status,
		ErrorMsg:      errorMsg,
		OperTime:      time.Now(),
	}
	body, err := json.Marshal(p)
	if err != nil {
		zlog.Warn().Err(err).Msg("操作日志序列化失败")
		return
	}
	// 构造 RocketMQ 消息：Topic 指向配置的操作日志 topic，Body 即序列化后的 payload
	msg := &primitive.Message{
		Topic: LogTopic,
		Body:  body,
	}
	stt, _ := time.ParseDuration(global.CFG.RocketMQ.SendTimeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), stt)
	defer cancel()
	if _, err := global.RocketMQ.SendSync(ctx, msg); err != nil {
		zlog.Warn().Err(err).Msg("操作日志入队失败")
	}
}

// maskedRequestParam 取 CaptureBody 缓存的请求体并脱敏（密码/令牌/密钥等字段打码）。
func maskedRequestParam(c *gin.Context) string {
	raw, ok := c.Get("requestBody")
	if !ok {
		return ""
	}
	b, ok := raw.([]byte)
	if !ok {
		return ""
	}
	return utils.MaskSensitiveJSON(b)
}
