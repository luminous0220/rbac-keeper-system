package config

import (
	"time"

	"github.com/spf13/viper"
)

// CFG 全局应用配置，由 viper 加载
type CFG struct {
	// Mode 运行模式：dev / prod，影响日志级别和中间件行为
	Mode string `mapstructure:"mode"`

	// LogMaxAge 旧日志文件保留天数，<=0 表示不清理旧日志
	LogMaxAge int `mapstructure:"log-max-age"`

	// Server HTTP 服务配置
	Server ServerConfig `mapstructure:"server"`

	// MySQL 文件元数据持久化配置
	MySQL MySQLConfig `mapstructure:"mysql"`

	// Redis 缓存/会话/消息队列/分布式锁配置
	Redis RedisConfig `mapstructure:"redis"`

	// MinIO 对象存储配置
	MinIO MinIOConfig `mapstructure:"minio"`

	// JWT 鉴权 Token 配置（双令牌：Access 短效 + Refresh 长效）
	JWT JWTConfig `mapstructure:"jwt"`

	// RateLimit 令牌桶限流配置（Redis 原子计数，分布式生效）
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`

	// RocketMQ 消息队列配置（替代原 Redis Stream，承载操作日志等异步消息）
	RocketMQ RocketMQConfig `mapstructure:"rocketmq"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	// Port 服务监听端口
	Port int `mapstructure:"port"`

	// ReadTimeout 请求体读取超时时间，包含整个请求体传输过程
	ReadTimeout time.Duration `mapstructure:"readTimeout"`

	// WriteTimeout 响应体写入超时时间，包含整个响应体传输过程
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

// MySQLConfig 文件元数据持久化配置
type MySQLConfig struct {
	// DSN 数据库连接字符串，格式：user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=true
	DSN string `mapstructure:"dsn"`

	// MaxOpenConns 最大打开连接数，0 表示无限制，建议设为 CPU 核心数的 2-4 倍
	MaxOpenConns int `mapstructure:"maxOpenConns"`

	// MaxIdleConns 最大空闲连接数，建议与 MaxOpenConns 保持一致或略小
	MaxIdleConns int `mapstructure:"maxIdleConns"`

	// ConnMaxLifetime 连接最大存活时间，超过后连接会被自动回收，建议设为 MySQL wait_timeout 的 70%
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`

	// LogLevel GORM 日志级别：0=Silent, 1=Error, 2=Warn, 3=Info, 4=Trace
	LogLevel int `mapstructure:"logLevel"`
}

// RedisConfig 缓存/会话/消息队列/分布式锁配置
type RedisConfig struct {
	// Addr Redis 服务地址，格式为 "host:port"
	Addr string `mapstructure:"addr"`

	// Password Redis 认证密码，为空表示无需认证
	Password string `mapstructure:"password"`

	// DB 选用的 Redis 数据库编号，范围 0-15
	DB int `mapstructure:"db"`

	// PoolSize 连接池大小，建议设为并发请求数的 1.5-2 倍
	PoolSize int `mapstructure:"poolSize"`
}

// MinIOConfig 对象存储配置
type MinIOConfig struct {
	// Endpoint MinIO 服务端点地址，格式为 "host:port"
	Endpoint string `mapstructure:"endpoint"`

	// AccessKey MinIO 访问密钥 ID
	AccessKey string `mapstructure:"accessKey"`

	// SecretKey MinIO 访问密钥密码
	SecretKey string `mapstructure:"secretKey"`

	// Bucket 默认存储桶名称
	Bucket string `mapstructure:"bucket"`

	// UseSSL 是否启用 SSL/TLS 加密连接
	UseSSL bool `mapstructure:"useSSL"`

	// PresignExpire 预签名 URL 有效期，过期后 URL 不可访问
	PresignExpire time.Duration `mapstructure:"presignExpire"`
}

// JWTConfig 鉴权 Token 配置（双令牌：Access 短效 + Refresh 长效）
type JWTConfig struct {
	// Secret JWT 签名密钥，建议使用 32 字节以上随机字符串
	Secret string `mapstructure:"secret"`

	// Issuer Token 签发者标识，用于校验 Token 来源
	Issuer string `mapstructure:"issuer"`

	// AccessExpire Access Token 有效期，建议 15-30 分钟，比如：30m
	AccessExpire time.Duration `mapstructure:"accessExpire"`

	// RefreshExpire Refresh Token 有效期，建议 7-30 天，比如：30d
	RefreshExpire time.Duration `mapstructure:"refreshExpire"`
}

// RateLimitConfig 令牌桶限流配置（Redis 原子计数，分布式生效）
type RateLimitConfig struct {
	// Rate 每秒生成的令牌数，即平均请求速率上限
	Rate int `mapstructure:"rate"`

	// Burst 令牌桶最大容量，允许突发请求超过 Rate 的瞬时峰值
	Burst int `mapstructure:"burst"`
}

// RocketMQConfig RocketMQ 消息队列配置（替代原 Redis Stream）
type RocketMQConfig struct {
	// Namesrv NameServer 地址列表，格式 "host:port"，多个用逗号分隔（例：127.0.0.1:9876,10.0.0.2:9876）
	Namesrv string `mapstructure:"namesrv"`
	// BrokerAddr Broker 地址 格式 "host:port"，多个用逗号分隔（例：127.0.0.1:9876,10.0.0.2:9876）
	BrokerAddr string `mapstructure:"brokerAddr"`
	// Instance 消费者实例名，同组内需唯一（用于区分同进程的多个消费者实例），留空由客户端自动生成
	Instance string `mapstructure:"instance"`

	// AccessKey/SecretKey ACL 访问密钥。未开启 RocketMQ ACL 时留空即可
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`

	// MaxReconsumeTimes 消费失败最大重试次数。超过后消息自动转入死信队列
	//（topic 形如 %DLQ%<group>），相当于原 Redis Stream 的手写死信 "<stream>:dlq"。
	// 默认 16，设为 -1 表示无限重试。
	MaxReconsumeTimes int32 `mapstructure:"maxReconsumeTimes"`

	// SendTimeout 生产者同步发送的超时时间（例：3s）。入队不阻塞业务太久
	SendTimeout string `mapstructure:"sendTimeout"`
}

// loadConfig 从指定路径读取配置文件，path 用于在测试中指向仓库根的 config.yaml
func LoadConfig(path string) (config CFG) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()

	// viper.Unmarshal 默认注册了 StringToTimeDurationHook，可直接将 "24h"/"10s" 解析为 time.Duration
	err = viper.Unmarshal(&config)

	if err != nil {
		panic(err)
	}
	return
}
