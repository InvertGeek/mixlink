package server

import (
	"fmt"
	"io"
	"log"
	"mixlink/internal/config"
	"mixlink/internal/database"
	"mixlink/internal/utils"
	"net/http"
	"time"
)

// 创建反向代理 handler
func proxyHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var target = config.MatchTarget(r.Host)
		remoteUrl := target.URL + r.RequestURI

		if config.Config.NoQuery {
			remoteUrl = target.URL + r.URL.Path
		}

		cachedRecord, _ := database.GetURL(remoteUrl)
		uploading := cachedRecord.IsUploading()

		//上传状态已经超过1分钟未更新，视为超时
		if uploading && time.Since(cachedRecord.CheckedTime) > time.Minute {
			_ = database.DeleteURL(remoteUrl)
			uploading = false
		}

		var shouldCacheByExt = config.ShouldCacheByExt(r.URL.Path)

		if shouldCacheByExt && HandleCachedURL(w, r, cachedRecord, remoteUrl) {
			return
		}

		proxyReq, err := http.NewRequest(r.Method, remoteUrl, r.Body)
		if err != nil {
			utils.HttpError(w, err, "创建请求失败")
			return
		}

		// 复制原请求头,go中host是特殊header,不在header列表中
		proxyReq.Header = r.Header.Clone()

		//自定义Host请求头
		targetHost := target.Host
		if targetHost != nil {
			proxyReq.Host = *targetHost
		}

		client := utils.Client
		remoteResponse, err := client.Do(proxyReq)
		defer remoteResponse.Body.Close()
		if err != nil {
			utils.HttpError(w, err, "请求目标服务器失败")
			return
		}

		var shouldCache = shouldCacheByExt && remoteResponse.ContentLength < (target.SizeLimit*1024)

		if !uploading && shouldCache {
			go func() {
				var referer = r.Referer()
				if err := HandleUpload(remoteUrl, r.URL.Path, referer); err != nil {
					log.Printf("上传文件失败:%v %v", remoteUrl, err)
					return
				}
			}()
		}

		utils.CopyHeaders(w.Header(), remoteResponse.Header)

		w.WriteHeader(remoteResponse.StatusCode)
		_, err = io.Copy(w, remoteResponse.Body)
		if err != nil {
			utils.HttpError(w, err, "转发请求体失败")
			return
		}
	}
}

func Start() {

	http.HandleFunc("/", proxyHandler())

	port, _ := utils.FindAvailablePort(config.Config.Port)
	addr := fmt.Sprintf("%v:%v", config.Config.Host, port)
	log.Printf("加速代理已启动，监听 %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
