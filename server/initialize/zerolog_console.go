package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"
)

// ANSI 转义颜色码常量，用于控制台日志整行着色
// 格式说明：\x1b[<样式>;<前景色>m 为设置样式，\x1b[0m 为重置所有样式
const (
	colorReset = "\x1b[0m"    // 重置所有样式与颜色
	colorDebug = "\x1b[35m"   // 洋红 Magenta
	colorInfo  = "\x1b[36m"   // 青色 Cyan
	colorWarn  = "\x1b[33m"   // 黄色 Yellow
	colorError = "\x1b[31m"   // 红色 Red
	colorFatal = "\x1b[1;31m" // 加粗红 Bold Red（fatal/panic）
)

// levelColor 返回日志级别对应的整行 ANSI 颜色码
func levelColor(level zerolog.Level) string {
	switch level {
	case zerolog.TraceLevel, zerolog.DebugLevel:
		return colorDebug
	case zerolog.InfoLevel:
		return colorInfo
	case zerolog.WarnLevel:
		return colorWarn
	case zerolog.ErrorLevel:
		return colorError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return colorFatal
	default:
		return colorReset
	}
}

// consoleWriter 彩色控制台写入器，实现 zerolog.LevelWriter 接口
// 接收 zerolog 生成的 JSON event，整行按级别着色后输出到 Out
// 设计：整行(时间/级别/文件行号/信息/字段)统一应用级别对应颜色，
// 而非 zerolog ConsoleWriter 默认的仅级别字段着色
type consoleWriter struct {
	out io.Writer
}

// newconsoleWriter 创建彩色控制台写入器
// out：底层输出目标，通常为 os.Stderr
func newConsoleWriter(out io.Writer) *consoleWriter {
	return &consoleWriter{out: out}
}

// field 有序字段，保留字段在 JSON 事件中的原始出现顺序。
type field struct {
	key string
	val interface{}
}

// parseOrdered 流式解析 JSON 对象为有序字段切片，保留字段原始出现顺序。
func parseOrdered(p []byte) ([]field, error) {
	// 创建一个基于字节切片 p 的 JSON 解码器，用于流式逐 token 解析
	dec := json.NewDecoder(bytes.NewReader(p))
	// dec.Token() 返回的类型是 interface{}，它可能是一个字符串（键）、一个数字、一个布尔值，也可能是一个 json.Delim（分隔符）
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	// json.Delim：这是 encoding/json 包中定义的一个类型，本质是 byte，专门用来表示 JSON 中的分隔符：{、}、[、]
	// 读取第一个 token，如果第一个支付不是起始分隔符 '{'，表示该 json 不合法
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("期望 JSON 对象")
	}

	var fields []field
	// 循环读取 JSON 对象内部的键值对，dec.More() 在遇到 '}' 时返回 false
	for dec.More() {
		// 获取 json 的 key
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		// / 声明一个空接口变量，用于接收任意类型的值
		var val interface{}
		// 解码当前键对应的值，自动推断类型并存入 val
		// 读取一个完整的 JSON 值。它可以递归地解析整个嵌套对象、数组或基本值
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		fields = append(fields, field{keyTok.(string), val})
	}
	// dec.Token() 的行为像 pop 一样，每次调用都会消费掉一个 token，解码器的内部指针会被推进。当所有 token 都被读取完毕后，再次调用就会返回 io.EOF
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return fields, nil
}

// isMeta 判断是否为 zerolog 元字段(时间/级别/调用者/信息)，需单独格式化
func isMeta(key string) bool {
	switch key {
	case zerolog.TimestampFieldName, zerolog.LevelFieldName,
		zerolog.CallerFieldName, zerolog.MessageFieldName:
		return true
	}
	return false
}

// WriteLevel 实现 zerolog.LevelWriter，按级别整行着色输出
// 参数 level：日志级别，决定整行颜色；p：zerolog 生成的 JSON event 字节流
// 返回值 n：原始字节长度(满足 io.Writer 契约)；err：写入错误
func (w *consoleWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	// zerolog 按链式调用顺序写入字段，如果这里使用 json.Unmarshal 解析到 map 后遍历顺序随机。即使原始 JSON 是 {"a":"1","b":"2","c":"3"}，解析后的 map 在遍历时可能输出 b, c, a，也可能输出 c, a, b，每次运行都可能不同
	// 所以使用 parseOrdered 流式解析 JSON 对象为有序字段切片，保留字段原始出现顺序
	fields, err := parseOrdered(p)
	// 解析失败则降级为原样输出(不带颜色)，保证日志不丢
	if err != nil {
		return w.out.Write(p)
	}

	// 预取元字段值(按名称定位，与其在事件中的位置无关)
	meta := make(map[string]interface{}, 4)
	for _, f := range fields {
		if isMeta(f.key) {
			meta[f.key] = f.val
		}
	}

	var sb strings.Builder
	// 整行颜色前缀：从此处开始整行应用级别颜色
	sb.WriteString(levelColor(level))

	// 时间
	if t, ok := meta[zerolog.TimestampFieldName].(string); ok && t != "" {
		sb.WriteString(t)
	}
	// 级别
	sb.WriteString(" ")
	sb.WriteString(strings.ToUpper(level.String()))
	// 文件&行号
	if c, ok := meta[zerolog.CallerFieldName].(string); ok && c != "" {
		sb.WriteString(" ")
		sb.WriteString(c)
	}
	// 日志信息
	if m, ok := meta[zerolog.MessageFieldName].(string); ok && m != "" {
		sb.WriteString(" >> ")
		sb.WriteString(m)
	}
	// 结构化字段(排除元字段)，按事件中原始出现顺序输出
	for _, f := range fields {
		if isMeta(f.key) {
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s=%v", f.key, f.val))
	}

	// 颜色重置 + 换行
	sb.WriteString(colorReset)
	sb.WriteString("\n")

	_, _ = w.out.Write([]byte(sb.String()))
	return len(p), nil
}

// Write 兜底实现 io.Writer，无级别信息时按 NoLevel 处理
func (w *consoleWriter) Write(p []byte) (int, error) {
	return w.WriteLevel(zerolog.NoLevel, p)
}
