package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"mixlink/internal/config"
	"mixlink/internal/database"
	"mixlink/internal/utils"
	"net/http"
	"time"
)

// UploadWithHeartbeat 执行上传，并在上传过程中每隔 interval 更新一次数据库时间
func UploadWithHeartbeat(
	recordURL string,
	name string,
	uploadEndpoint string,
	uploadStream io.Reader,
	size int64,
	interval time.Duration,
) (string, error) {
	done := make(chan struct{})
	defer close(done)

	// 启动心跳 goroutine
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 更新数据库时间
				err := database.UpdateTime(recordURL)
				if err != nil {
					log.Println("更新记录失败:", err)
				}
			case <-done:
				return
			}
		}
	}()

	// 执行上传
	link, err := utils.CommonPutUpload(name, uploadEndpoint, size, uploadStream)
	if err != nil {
		return "", err
	}
	return link, nil
}

func HandleUpload(remoteUrl, requestPath string) error {
	remoteResponse, err := utils.Client.Get(remoteUrl)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	if remoteResponse.StatusCode < 200 || remoteResponse.StatusCode >= 400 {
		return nil
	}
	// 尝试添加 URL，避免重复上传
	if !database.AddURL(remoteUrl, "uploading") {
		return nil
	}

	log.Printf("开始上传文件: %v", remoteUrl)

	// 根据 Content-Encoding 选择上传流
	uploadStream := remoteResponse.Body
	if remoteResponse.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(remoteResponse.Body)
		if err != nil {
			return fmt.Errorf("创建 gzip reader 失败: %w", err)
		}
		defer gz.Close()
		uploadStream = gz
	}

	// 上传文件
	link, err := UploadWithHeartbeat(remoteUrl, requestPath, config.Config.UploadEndpoint, uploadStream, remoteResponse.ContentLength, 10*time.Second)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	// 上传成功后更新数据库
	database.AddURL(remoteUrl, link)
	log.Printf("上传文件成功: %v -> %v", remoteUrl, link)
	return nil
}

// HandleCachedURL 检查缓存记录是否可用，并处理跳转或失效
func HandleCachedURL(w http.ResponseWriter, r *http.Request, cachedRecord *database.URLRecord, remoteUrl string) (handled bool) {
	if cachedRecord == nil || cachedRecord.IsUploading() {
		return false
	}

	if cachedRecord.InValid >= config.Config.Invalid {
		_ = database.DeleteURL(remoteUrl)
		return false
	}

	var referer = r.Referer()
	// 检查链接有效性
	valid := time.Since(cachedRecord.Time) < time.Second || utils.HeadUrl(cachedRecord.Link, referer, 3*time.Second)
	if valid {
		_ = database.UpdateTime(remoteUrl)
		_ = database.ResetInvalid(remoteUrl)
		http.Redirect(w, r, cachedRecord.Link, http.StatusTemporaryRedirect)
		return true
	}

	// 无效则自增 InValid
	_ = database.IncInvalid(remoteUrl)
	return false
}
