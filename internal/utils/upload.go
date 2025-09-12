package utils

import (
	"fmt"
	"io"
	"net/http"
)

func CommonPutUpload(name string, putUrl string, size int64, body io.Reader) (string, error) {
	result, err := PutUpload(body, putUrl+"?name="+name, size, map[string]string{})
	if err != nil {
		return "", err
	}
	return result, nil
}

// PutUpload 执行HTTP PUT请求上传数据流
// stream: 要上传的数据流
// url: 目标URL
// headers: 自定义HTTP headers
func PutUpload(stream io.Reader, link string, size int64, headers map[string]string) (string, error) {
	// 创建HTTP客户端
	client := Client

	// 创建PUT请求
	req, err := http.NewRequest("PUT", link, stream)
	req.ContentLength = size
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置自定义headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
