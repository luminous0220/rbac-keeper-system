package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"rbackeeper/global"
	"rbackeeper/internal/consumer"
	"rbackeeper/internal/dao"
	"rbackeeper/internal/dto"
	"rbackeeper/internal/model"
	"rbackeeper/utils"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LoginLogList 登录日志分页
func LoginLogList(query dto.LoginLogListQuery) (*dto.LoginLogListResp, error) {
	rows, total, err := dao.GetLoginLogList(query)
	if err != nil {
		return nil, err
	}
	list := make([]dto.LoginLogItem, 0, len(rows))
	for i := range rows {
		list = append(list, toLoginLogItem(&rows[i]))
	}
	return &dto.LoginLogListResp{
		Total:    total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize,
		List:     list,
	}, nil
}

// OperationLogList 操作日志分页
func OperationLogList(query dto.OperationLogListQuery) (*dto.OperationLogListResp, error) {
	rows, total, err := dao.GetOperationLogList(query)
	if err != nil {
		return nil, err
	}
	list := make([]dto.OperationLogItem, 0, len(rows))
	for i := range rows {
		list = append(list, toOperationLogItem(&rows[i]))
	}
	return &dto.OperationLogListResp{
		Total:    total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize,
		List:     list,
	}, nil
}

// LogExport 导出登录/操作日志为 CSV；返回字节流与建议文件名。
// 默认导出近 7 天；超过 10 万行时强制截断（防 OOM）。
func LogExport(typ string) ([]byte, string, error) {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	switch typ {
	case "login":
		rows, err := dao.StreamLoginLogs(start, now)
		if err != nil {
			return nil, "", err
		}
		buf := &bytes.Buffer{}
		w := csv.NewWriter(buf)
		_ = w.Write([]string{"ID", "用户名", "IP", "归属", "浏览器", "系统", "状态", "消息", "登录时间"})
		for i := range rows {
			if i >= 100000 {
				break
			}
			r := &rows[i]
			_ = w.Write([]string{
				fmt.Sprintf("%d", r.ID), r.Username, r.IP, r.IPAddr, r.Browser, r.OS,
				statusText(r.Status), r.Msg, r.LoginTime.Format("2006-01-02 15:04:05"),
			})
		}
		w.Flush()
		return buf.Bytes(), "login_logs.csv", nil
	case "operation":
		rows, err := dao.StreamOperationLogs(start, now)
		if err != nil {
			return nil, "", err
		}
		buf := &bytes.Buffer{}
		w := csv.NewWriter(buf)
		_ = w.Write([]string{"ID", "模块", "方法", "操作人", "URL", "IP", "状态", "错误", "操作时间"})
		for i := range rows {
			if i >= 100000 {
				break
			}
			r := &rows[i]
			_ = w.Write([]string{
				fmt.Sprintf("%d", r.ID), r.Title, r.RequestMethod,
				r.OperName, r.OperURL, r.OperIP,
				statusText(r.Status), r.ErrorMsg, r.OperTime.Format("2006-01-02 15:04:05"),
			})
		}
		w.Flush()
		return buf.Bytes(), "operation_logs.csv", nil
	default:
		return nil, "", fmt.Errorf("unknown log type: %s", typ)
	}
}

func toLoginLogItem(r *model.LoginLog) dto.LoginLogItem {
	return dto.LoginLogItem{
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

func toOperationLogItem(r *model.OperationLog) dto.OperationLogItem {
	return dto.OperationLogItem{
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

// 异步入队一条操作日志。失败仅 Warn，不阻断业务。
// 用法：controller 业务成功 return 前调一次；越权 403 时也调用（status=1）。
func CreateLogWithMq(c *gin.Context, title string) {
	// 生产者未初始化（如未配置 RocketMQ 直接启动）时静默跳过，不阻断业务
	if global.RocketMQ == nil || global.CFG == nil {
		return
	}
	claims := utils.GetClaims(c)
	operName := ""
	if claims != nil {
		operName = claims.Username
	}
	p := consumer.OperationLogPayload{
		Title:         title,
		RequestMethod: c.Request.Method,
		OperName:      operName,
		OperURL:       c.Request.URL.Path,
		OperIP:        c.ClientIP(),
		OperParam:     maskedRequestParam(c),
		Status:        0,
		OperTime:      time.Now(),
	}
	body, err := json.Marshal(p)
	if err != nil {
		log.Warn().Err(err).Msg("操作日志序列化失败")
		return
	}
	// 构造 RocketMQ 消息：Topic 指向配置的操作日志 topic，Body 即序列化后的 payload
	msg := &primitive.Message{
		Topic: consumer.LogTopic,
		Body:  body,
	}
	stt, _ := time.ParseDuration(global.CFG.RocketMQ.SendTimeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), stt)
	defer cancel()
	if _, err := global.RocketMQ.SendSync(ctx, msg); err != nil {
		log.Warn().Err(err).Msg("操作日志入队失败")
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
