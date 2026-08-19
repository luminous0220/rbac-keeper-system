package utils

import (
	"net"
	"strings"
)

// 解析 IP 归属（最小实现：内网/未知）。
// 完整 IP 库（ip2region / qqwry）作为后续可选项，本期不做。
func ParseIPLocation(ip string) string {
	if ip == "" {
		return "未知地区"
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "未知地区"
	}
	// 环回 / 链路本地 / 私网
	if parsed.IsLoopback() || parsed.IsLinkLocalUnicast() || parsed.IsPrivate() {
		return "内网IP"
	}
	// IPv6 段（除 loopback）
	if strings.Contains(ip, ":") {
		return "IPv6地址"
	}
	return "未知地区"
}
