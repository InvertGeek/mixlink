package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
)

// Run 接收一个无参数的函数并立即执行，返回函数的返回值
func Run[T any](f func() T) T {
	return f()
}

func ParseUrl(text string) string {
	parsed, _ := url.Parse(text)
	return parsed.String()
}

func RemoveQueryFromURL(rawURL string) (string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 清空 query 参数
	parsedURL.RawQuery = ""

	// 返回不带 query 的 URL
	return parsedURL.String(), nil
}

func DecompressGzip(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func GenRandomHexString(length int) string {
	rBytes := make([]byte, length/2)
	rand.Read(rBytes)
	return hex.EncodeToString(rBytes)
}

func FindAvailablePort(startPort int) (int, error) {
	for port := startPort; port <= 65535; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			err := listener.Close()
			if err != nil {
				return 0, err
			} // 关闭监听器，释放端口
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found")
}

// GetHost 从字符串形式的 URL 中提取 Host
func GetHost(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// GetPath 从字符串形式的 URL 中提取 Path
func GetPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Path, nil
}
