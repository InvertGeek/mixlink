
## 基于反向代理的静态资源加速

### 原理
- 用户访问站点时，会转发原站返回内容，相当于一个基础的反向代理服务。  
- 当访问的 URL 命中特定后缀时，会将原站资源上传至加速服务。  
- 下次访问相同资源时，直接重定向到加速地址，不消耗服务器流量。  

### 部署
1. 需要先部署上传服务，推荐使用 [directlink](https://github.com/InvertGeek/directlink)。  
2. 上传服务接口填写为：`directlink服务地址 + /api/upload`。  
3. 运行后编辑 `config.yaml` 文件，自行配置即可。 

### 自定义过期时间
可自定义缓存过期时间 \
原站响应携带x-mixlink-expire响应头即可指定缓存过期时间 \
格式为毫秒,过期后会删除缓存重新上传,默认永不过期

### 链接失效
- 每次重定向之前会检测加速链接是否可访问。  
- 如果不可访问，将会 **fallback 到转发模式**，不用担心失效问题。  
- 同一 URL 资源连续多次检测到失效时，会自动删除并重新上传。  

### 已知问题
css如果存在相对路径引用，会导致路径错误 \
js esm语法相对路径引用也会导致路径错误,webpack等工具可把esm编译为commonjs\
因为是相对于文件本身路径，而不是html页面路径 

---

## GitHub Star
[![Stargazers over time](https://starchart.cc/InvertGeek/mixlink.svg?variant=adaptive)](https://starchart.cc/InvertGeek/mixlink)
