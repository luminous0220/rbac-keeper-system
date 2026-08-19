package consumer

import (
	"context"
	"encoding/json"
	"rbackeeper/global"
	"rbackeeper/initialize"
	"rbackeeper/internal/model"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/rs/zerolog/log"
)

const LogTopic = "LOG"

// OperationLogPayload 操作日志队列 payload（消费者侧反序列化用）
type OperationLogPayload struct {
	Title         string    `json:"title"`
	RequestMethod string    `json:"requestMethod"`
	OperName      string    `json:"operName"`
	OperURL       string    `json:"operUrl"`
	OperIP        string    `json:"operIp"`
	OperParam     string    `json:"operParam"`
	Status        int8      `json:"status"`
	ErrorMsg      string    `json:"errorMsg"`
	OperTime      time.Time `json:"operTime"`
}

// 初始化消费者
func InitOperationLog() {
	initialize.RegisterConsumer(&initialize.ConsumerConfig{
		Topic:   LogTopic,
		Group:   "log_consumer_group",
		Handler: handler,
	})
}

// 处理单条操作日志消息：反序列化 payload → 落库。
// 返回 nil 表示成功（将 ACK）；返回 error 表示失败（触发重投，超限进死信队列）。
func handler(ctx context.Context, msg *primitive.MessageExt) error {
	var p OperationLogPayload
	if err := json.Unmarshal(msg.Body, &p); err != nil {
		// payload 损坏不会随重投自愈，这里仍返回错误交给 RocketMQ：
		// 重试满 maxReconsumeTimes 次后自动沉入死信队列，便于人工排查
		log.Warn().Err(err).Str("msg_id", msg.MsgId).Msg("操作日志 payload 解析失败")
		return err
	}

	row := &model.OperationLog{
		Title:         p.Title,
		RequestMethod: p.RequestMethod,
		OperatorType:  1,
		OperName:      p.OperName,
		OperURL:       p.OperURL,
		OperIP:        p.OperIP,
		OperParam:     p.OperParam,
		Status:        p.Status,
		ErrorMsg:      p.ErrorMsg,
		OperTime:      p.OperTime,
	}

	if err := global.DB.Create(row).Error; err != nil {
		// ReconsumeTimes 表示已重投次数，打出来便于监控重试进度
		log.Warn().Err(err).Str("msg_id", msg.MsgId).
			Int32("reconsume_times", msg.ReconsumeTimes).Msg("操作日志落库失败，等待重投")
		return err
	}
	return nil
}
