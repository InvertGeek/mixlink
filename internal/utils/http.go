package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

var Client = &http.Client{}

// CopyHeaders 将源 Header 原封不动复制到目标 Header
func CopyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// HeadUrl 检测 URL 是否有效，可指定 Referer
func HeadUrl(url, referer string, timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}

	// 使用 http.NewRequest 设置 Referer
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	// 设置 Referer 头
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 只要状态码是 2xx 或 3xx 就认为有效
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func HttpError(w http.ResponseWriter, err error, msg string) {
	fullMsg := fmt.Sprintf("%s: %v\n%s", msg, err, debug.Stack())
	log.Printf(fullMsg)
	http.Error(w, fullMsg, http.StatusBadGateway)
}

// IsURLAlive 检测 URL 是否有效，兼容 HEAD/Range 都不支持的服务器
func IsURLAlive(url string, timeout time.Duration) bool {
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 立即关闭 Body，不必读取全部内容
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true
	}

	return false
}
