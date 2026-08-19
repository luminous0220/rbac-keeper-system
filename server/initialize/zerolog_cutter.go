package initialize

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// cleanInterval 两次过期日志清理之间的最小间隔，避免高频写入时频繁扫描目录
const cleanInterval = time.Hour

// fileCutter 按日期 + 级别分文件的日志切割器，实现 zerolog.LevelWriter 接口。
// 目录结构：<dir>/<年>/<月-日>/<级别>.log，例如 logs/2026/01-01/info.log。
// 跨天自动切换到新日期目录下的新文件；可选按保留天数清理过期日志目录。
type fileCutter struct {
	mu        sync.Mutex                   // 保护 writers 懒创建与 lastClean 去重
	dir       string                       // 日志根目录，如 "logs"
	maxAge    int                          // 旧日志保留天数，<=0 表示不清理
	lastClean time.Time                    // 上次清理时间，用于去重
	writers   map[zerolog.Level]*dailyFile // 各级别对应的日文件（懒加载）
}

// dailyFile 单个级别在某一天的文件句柄，负责按天切换文件
type dailyFile struct {
	mu      sync.Mutex // 保护 file/curDate 的并发切换与写入
	dir     string     // 日志根目录
	name    string     // 级别文件名，如 "info.log"
	file    *os.File   // 当前打开的文件句柄
	curDate string     // 当前文件所属日期路径，如 "2026/01-01"
}

// newFileCutter 创建按日期 + 级别分文件的日志切割器
// dir：日志根目录；maxAge：旧日志保留天数（<=0 不清理）
func newFileCutter(dir string, maxAge int) *fileCutter {
	return &fileCutter{
		dir:     dir,
		maxAge:  maxAge,
		writers: make(map[zerolog.Level]*dailyFile),
	}
}

// datePath 返回给定时间对应的相对日期路径，如 "2026/01-01"（年/月-日）
func datePath(t time.Time) string {
	return t.Format("2006/01-02")
}

// levelFileName 返回级别对应的文件名，如 info.log；NoLevel 等无级别字符串兜底为 log.log
func levelFileName(level zerolog.Level) string {
	name := level.String()
	if name == "" {
		name = "log"
	}
	return name + ".log"
}

// WriteLevel 实现 zerolog.LevelWriter，按级别路由到对应日文件
// 参数 level：日志级别；p：zerolog 生成的完整日志字节流（一行 JSON）
// 返回值 n：写入字节数；err：写入错误
func (w *fileCutter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	// 仅对 writers map 加锁保护懒创建；dailyFile.write 内部自持锁，无需持锁写入
	w.mu.Lock()
	df := w.writers[level]
	if df == nil {
		df = &dailyFile{
			dir:  w.dir,
			name: levelFileName(level),
		}
		w.writers[level] = df
	}
	w.mu.Unlock()

	w.maybeClean()
	return df.write(p)
}

// Write 兜底实现 io.Writer，无级别信息时按 NoLevel 处理，写入 log.log
func (w *fileCutter) Write(p []byte) (int, error) {
	return w.WriteLevel(zerolog.NoLevel, p)
}

// maybeClean 周期性触发过期日志目录清理，两次清理间隔不小于 cleanInterval
// 清理在独立 goroutine 异步执行，不阻塞日志写入
func (w *fileCutter) maybeClean() {
	w.mu.Lock()
	if w.maxAge <= 0 {
		w.mu.Unlock()
		return
	}
	if !w.lastClean.IsZero() && time.Since(w.lastClean) < cleanInterval {
		w.mu.Unlock()
		return
	}
	w.lastClean = time.Now()
	w.mu.Unlock()

	go w.cleanOld(time.Now())
}

// cleanOld 扫描日期目录，删除超过保留天数的日志目录
// 目录命名：<dir>/<年>/<月-日>/，按 "2006/01-02" 解析为日期与截止时间比较
func (w *fileCutter) cleanOld(now time.Time) {
	cutoff := now.AddDate(0, 0, -w.maxAge)
	years, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}
	for _, y := range years {
		if !y.IsDir() {
			continue
		}
		yearDir := filepath.Join(w.dir, y.Name())
		days, err := os.ReadDir(yearDir)
		if err != nil {
			continue
		}
		for _, day := range days {
			if !day.IsDir() {
				continue
			}
			t, err := time.ParseInLocation("2006/01-02", y.Name()+"/"+day.Name(), time.Local)
			if err != nil {
				continue
			}
			if t.Before(cutoff) {
				_ = os.RemoveAll(filepath.Join(yearDir, day.Name()))
			}
		}
	}
}

// write 写入字节流，必要时按当天日期切换文件
func (d *dailyFile) write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	today := datePath(now)
	// 日期变更或首次写入时切换到当天文件
	if d.curDate != today || d.file == nil {
		if err := d.rotate(today); err != nil {
			return 0, err
		}
	}
	return d.file.Write(p)
}

// rotate 切换到指定日期对应的文件：关闭旧句柄、创建日期目录、打开新文件（追加模式）
func (d *dailyFile) rotate(today string) error {
	if d.file != nil {
		_ = d.file.Close()
		d.file = nil
	}
	dir := filepath.Join(d.dir, today)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, d.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	d.file = f
	d.curDate = today
	return nil
}
