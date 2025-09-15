package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"mixlink/internal/config"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

var Client = &http.Client{
	Timeout: config.Config.MaxTimeout,
	Transport: &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			proxy := strings.TrimSpace(config.Config.ProxyUrl)
			if proxy != "" {
				return url.Parse(proxy)
			}
			return http.ProxyFromEnvironment(req)
		},
	},
}

// DoRequest 用固定的 Client 发请求，支持超时控制
func DoRequest(req *http.Request, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	return Client.Do(req) // 👈 固定全局 Client
}

// CopyHeaders 将源 Header 原封不动复制到目标 Header
func CopyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// HeadUrl 检测 URL 是否有效，可指定 Referer
func HeadUrl(targetURL, referer string, timeout time.Duration) bool {
	req, err := http.NewRequest(http.MethodHead, targetURL, nil)
	if err != nil {
		return false
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := DoRequest(req, timeout)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func HeadersToMap(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for name, values := range h {
		// 多个值用逗号拼接
		m[name] = strings.Join(values, ",")
	}
	return m
}

func HttpError(w http.ResponseWriter, err error, msg string) {
	fullMsg := fmt.Sprintf("%s: %v\n%s", msg, err, debug.Stack())
	log.Printf(fullMsg)
	http.Error(w, fullMsg, http.StatusBadGateway)
}

// IsURLAlive 检测 URL 是否有效，兼容 HEAD/Range 都不支持的服务器
func IsURLAlive(targetURL string, timeout time.Duration) bool {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return false
	}

	resp, err := DoRequest(req, timeout)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
