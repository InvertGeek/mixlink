package server

import (
	"compress/gzip"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"io"
	"log"
	"mixlink/internal/config"
	"mixlink/internal/database"
	"mixlink/internal/utils"
	"net/http"
	"time"
)

var sem = semaphore.NewWeighted(config.Config.UploadTask)

// UploadWithHeartbeat 执行上传，并在上传过程中每隔 interval 更新一次数据库时间
func UploadWithHeartbeat(
	recordURL string,
	name string,
	uploadEndpoint string,
	uploadStream io.Reader,
	size int64,
	interval time.Duration,
) (string, error) {
	stopHeartbeat := utils.StartHeartbeat(interval, func() {
		// 这里写自定义逻辑，比如更新数据库
		if err := database.UpdateTime(recordURL); err != nil {
			log.Println("更新记录失败:", err)
		}
	})
	defer stopHeartbeat()

	if err := sem.Acquire(context.Background(), 1); err != nil {
		fmt.Println("acquire error:", err)
		return "", err
	}

	log.Printf("开始上传文件: %v", recordURL)

	defer sem.Release(1) // 用完释放

	// 执行上传
	link, err := utils.CommonPutUpload(name, uploadEndpoint, size, uploadStream)
	if err != nil {
		return "", err
	}
	return link, nil
}

func HandleUpload(remoteUrl, requestPath, referer string) error {
	// 构建请求
	req, err := http.NewRequest("GET", remoteUrl, nil)
	if err != nil {
		return err
	}
	// 设置 Referer
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	// 发请求
	remoteResponse, err := utils.Client.Do(req)
	if err != nil {
		return err
	}
	if remoteResponse.StatusCode < 200 || remoteResponse.StatusCode >= 400 {
		return nil
	}
	// 尝试添加 URL，避免重复上传
	if !database.AddURL(remoteUrl, "uploading", -1, -1) {
		return nil
	}

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

	var expireHeader = remoteResponse.Header.Get("x-mixlink-expire")
	var expire = utils.StrToInt64(expireHeader, -1)

	// 上传文件
	link, err := UploadWithHeartbeat(remoteUrl, requestPath, config.Config.UploadEndpoint, uploadStream, remoteResponse.ContentLength, 10*time.Second)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	// 上传成功后更新数据库
	database.AddURL(remoteUrl, link, remoteResponse.ContentLength, expire)
	log.Printf("上传文件成功: %v -> %v", remoteUrl, link)
	return nil
}

// HandleCachedURL 检查缓存记录是否可用，并处理跳转或失效
func HandleCachedURL(w http.ResponseWriter, r *http.Request, cachedRecord *database.URLRecord, remoteUrl string) (handled bool) {
	if cachedRecord == nil || cachedRecord.IsUploading() {
		return false
	}

	if cachedRecord.InValidTimes >= config.Config.Invalid || cachedRecord.IsExpired() {
		_ = database.DeleteURL(remoteUrl)
		return false
	}

	var referer = r.Referer()

	if !cachedRecord.CheckValid(referer) {
		log.Printf("文件失效: %v", cachedRecord.Link)
		_ = database.IncInvalid(remoteUrl)
		return false
	}

	if config.Config.LogRequest {
		log.Printf("重定向: %v -> %v", remoteUrl, cachedRecord.Link)
	}

	_ = database.UpdateTime(remoteUrl)
	_ = database.ResetInvalid(remoteUrl)
	http.Redirect(w, r, cachedRecord.Link, http.StatusTemporaryRedirect)
	return true
}
