package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/time/rate"
)

// 1. 定义一个全局的日志采样限流器
// 每秒最多允许打印 5 条日志，桶容量也是 5
// 这意味着：无论并发量多大，日志输出速率被死死限制在 5条/秒
var logLimiter = rate.NewLimiter(5, 5)

// 2. 封装一个采样日志打印函数
func sampledLog(format string, v ...interface{}) {
	// 核心：非阻塞检查是否有令牌
	if logLimiter.Allow() {
		// 拿到令牌，正常打印
		log.Printf(format, v...)
	}
}

func Test() {
	// 模拟一个高并发的请求处理过程
	// 假设我们在 1 秒内瞬间处理了 20 个请求
	fmt.Println("=== 模拟 20 个并发请求 ===")
	for i := 1; i <= 20; i++ {
		// 模拟业务处理
		processRequest(i)
	}

	// 等待一段时间，让令牌桶补充
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 2秒后，再次模拟 10 个并发请求 ===")
	for i := 21; i <= 30; i++ {
		processRequest(i)
	}
}

func processRequest(id int) {
	// 假设每个请求都会触发一条日志
	// 使用我们的采样日志函数，而不是 log.Printf
	sampledLog("处理请求 ID: %d", id)
}
