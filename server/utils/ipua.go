package utils

import "strings"

// 从 User-Agent 提取浏览器与操作系统（最小实现，IP 归属解析留待日志阶段）
func ParseUA(ua string) (browser, os string) {
	switch {
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	}
	switch {
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		os = "iOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "Mac OS"):
		os = "macOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	}
	return
}
