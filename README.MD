
## 基于反向代理的静态资源加速

### 原理
- 用户访问站点时，会转发原站返回内容，相当于一个基础的反向代理服务。  
- 当访问的 URL 命中特定后缀时，会将原站资源上传至加速服务。  
- 下次访问相同资源时，直接重定向到加速地址，不消耗服务器流量。  

### 部署
1. 需要先部署上传服务，推荐使用 [directlink](https://github.com/InvertGeek/directlink)。  
2. 上传服务接口填写为：`directlink服务地址 + /api/upload`。  
3. 运行后编辑 `config.yaml` 文件，自行配置即可。  

### 链接失效
- 每次重定向之前会检测加速链接是否可访问。  
- 如果不可访问，将会 **fallback 到转发模式**，不用担心失效问题。  
- 同一 URL 资源连续多次检测到失效时，会自动删除并重新上传。  

---

## GitHub Star
[![Stargazers over time](https://starchart.cc/InvertGeek/mixlink.svg?variant=adaptive)](https://starchart.cc/InvertGeek/mixlink)
