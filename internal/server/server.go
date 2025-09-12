package server

import (
	"fmt"
	"io"
	"log"
	"mixlink/internal/config"
	"mixlink/internal/database"
	"mixlink/internal/utils"
	"net/http"
	"net/url"
	"time"
)

// 创建反向代理 handler
func proxyHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var target = config.MatchTarget(r.Host)
		targetURL, err := url.Parse(target.URL)
		if err != nil {
			log.Fatalf("解析目标URL失败: %v", err)
		}
		remoteUrl := target.URL + r.RequestURI
		cachedRecord, _ := database.GetURL(remoteUrl)
		uploading := cachedRecord.IsUploading()

		//上传状态已经超过1分钟未更新，视为超时
		if uploading && time.Since(cachedRecord.Time) > time.Minute {
			_ = database.DeleteURL(remoteUrl)
			uploading = false
		}

		if handled := HandleCachedURL(w, r, cachedRecord, remoteUrl); handled {
			return
		}

		proxyReq, err := http.NewRequest(r.Method, remoteUrl, r.Body)
		if err != nil {
			utils.HttpError(w, err, "创建请求失败")
			return
		}

		// 复制原请求头，但 Host 由目标服务器决定
		proxyReq.Header = r.Header.Clone()
		proxyReq.Host = targetURL.Host

		client := utils.Client
		remoteResponse, err := client.Do(proxyReq)
		defer remoteResponse.Body.Close()
		if err != nil {
			utils.HttpError(w, err, "请求目标服务器失败")
			return
		}

		var shouldCache = config.ShouldCacheByExt(r.RequestURI) && remoteResponse.ContentLength < target.SizeLimit

		if !uploading && shouldCache {
			go func() {
				if err := HandleUpload(remoteUrl, r.RequestURI); err != nil {
					log.Printf("上传文件失败: %v", err)
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
	log.Printf("加速代理已启动启动，监听 %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
