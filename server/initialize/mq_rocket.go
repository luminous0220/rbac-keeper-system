package initialize

import (
	"context"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/rs/zerolog/log"

	"rbackeeper/config"
	"rbackeeper/global"
)

type ConsumerConfig struct {
	Topic   string
	Group   string
	Handler MessageHandler
}

var (
	consumers []*ConsumerConfig
)

type MessageHandler func(ctx context.Context, msg *primitive.MessageExt) error

// 注册消费者配置
func RegisterConsumer(cfg *ConsumerConfig) {
	consumers = append(consumers, cfg)
}

// 统一启动所有消费者
func StartAllConsumer(ctx context.Context) {
	for _, cfg := range consumers {
		go startSingleConsumer(ctx, cfg)
	}
}

func startSingleConsumer(ctx context.Context, cfg *ConsumerConfig) {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(ParseNamesrv(global.CFG.RocketMQ.Namesrv)),
		consumer.WithGroupName(cfg.Group),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	)
	if err != nil {
		log.Fatal().Err(err).Str("topic", cfg.Topic).Msg("MQ Consumer 创建失败")
	}

	err = c.Subscribe(cfg.Topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			if err := cfg.Handler(ctx, msg); err != nil {
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		log.Fatal().Err(err).Str("topic", cfg.Topic).Msg("MQ Consumer 订阅失败")
	}
	if err := c.Start(); err != nil {
		log.Fatal().Err(err).Str("topic", cfg.Topic).Msg("MQ Consumer 启动失败")
	}

	// 优雅退出
	go func() {
		<-ctx.Done()
		_ = c.Shutdown()
	}()
	log.Info().Str("topic", cfg.Topic).Str("group", cfg.Group).Msg("MQ Consumer 启动成功")
}

// 返回所有已注册消费者去重后的 topic 列表，供启动前 EnsureTopics 兜底建 topic
func Topics() []string {
	seen := make(map[string]struct{}, len(consumers))
	topics := make([]string, 0, len(consumers))
	for _, cfg := range consumers {
		if _, ok := seen[cfg.Topic]; ok {
			continue
		}
		seen[cfg.Topic] = struct{}{}
		topics = append(topics, cfg.Topic)
	}
	return topics
}

// 确保 topic 已存在，不存在则直接建到 broker 上。
// RocketMQ 的 topic 默认由「生产者首次发消息」触发自动创建；但消费者启动时
// 必须已能拿到该 topic 的路由，否则直接报 "topic not exist" 并退出。
// 因此在启动消费者前调用本函数兜底，避免每个新 topic 都要手工建一次。
func EnsureConsumerTopics(ctx context.Context, topics ...string) {

	if global.CFG.RocketMQ.Namesrv == "" {
		log.Fatal().Msg("RocketMQ Namesrv 未配置")
	}

	if global.CFG.RocketMQ.BrokerAddr == "" {
		log.Fatal().Msg("RocketMQ BrokerAddr 未配置")
	}

	if len(topics) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := []admin.AdminOption{
		admin.WithResolver(primitive.NewPassthroughResolver(ParseNamesrv(global.CFG.RocketMQ.Namesrv))),
	}

	AccessKey := global.CFG.RocketMQ.AccessKey
	SecretKey := global.CFG.RocketMQ.SecretKey

	if AccessKey != "" && SecretKey != "" {
		opts = append(opts, admin.WithCredentials(primitive.Credentials{
			AccessKey: AccessKey,
			SecretKey: SecretKey,
		}))
	}

	cli, err := admin.NewAdmin(opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("MQ Admin 客户端创建失败")
	}
	defer cli.Close()

	// 拉取 namesrv 上已有的全量 topic
	list, err := cli.FetchAllTopicList(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("拉取 RocketMQ topic 列表失败")
	}
	exists := make(map[string]struct{}, len(list.TopicList))
	for _, t := range list.TopicList {
		exists[t] = struct{}{}
	}

	for _, topic := range topics {
		if _, ok := exists[topic]; ok {
			log.Debug().Str("topic", topic).Msg("topic 已存在，跳过创建")
			continue
		}
		if err := cli.CreateTopic(ctx,
			admin.WithTopicCreate(topic),
			admin.WithBrokerAddrCreate(global.CFG.RocketMQ.BrokerAddr),
		); err != nil {
			log.Fatal().Err(err).Str("topic", topic).Msg("自动创建 topic 失败")
		}
		log.Info().Str("topic", topic).Msg("自动创建 topic 成功")
	}
}

// 初始化 RocketMQ 生产者并写入全局单例 global.RocketMQ。
//   - 配置 ACL 时挂载 Credentials；未配置则客户端走无鉴权通道。
//   - 客户端内部日志默认打到 stdout（logrus），这里把级别调到 warn 减少刷屏；
func RegisterRocketMQ(cfg config.CFG) {
	if cfg.RocketMQ.Namesrv == "" {
		log.Fatal().Msg("RocketMQ namesrv 未配置")
	}

	// 客户端日志级别：warn 起避免启动时大量 info 噪音
	rlog.SetLogLevel("warn")

	stt, _ := time.ParseDuration(cfg.RocketMQ.SendTimeout)
	opts := []producer.Option{
		// NameServer 地址列表（生产建议 ≥2 个做高可用）
		producer.WithNameServer(ParseNamesrv(cfg.RocketMQ.Namesrv)),
		// 同步发送失败时的本地重试次数（RocketMQ 网络异常会换 Queue 重试）
		producer.WithRetry(3),
		// 单次发送超时，避免长时间阻塞请求 goroutine
		producer.WithSendMsgTimeout(stt),
	}
	if cfg.RocketMQ.AccessKey != "" && cfg.RocketMQ.SecretKey != "" {
		opts = append(opts, producer.WithCredentials(primitive.Credentials{
			AccessKey: cfg.RocketMQ.AccessKey,
			SecretKey: cfg.RocketMQ.SecretKey,
		}))
	}

	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("RocketMQ 生产者创建失败")
	}
	// Start 会向 NameServer 拉取路由、建立与 Broker 的长连接
	if err := p.Start(); err != nil {
		log.Fatal().Err(err).Msg("RocketMQ 生产者启动失败")
	}

	global.RocketMQ = p
	log.Info().Str("namesrv", cfg.RocketMQ.Namesrv).Msg("注册 RocketMQ 生产者成功")
}

// 将配置中的逗号分隔字符串 "a:9876,b:9876" 解析为 primitive.NamesrvAddr（[]string 别名）。
// 同时做 TrimSpace 与空项过滤，容忍手写配置的空白/尾逗号。
func ParseNamesrv(s string) primitive.NamesrvAddr {
	var addrs []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			addrs = append(addrs, v)
		}
	}
	return primitive.NamesrvAddr(addrs)
}
