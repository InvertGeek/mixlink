
## 基于反向代理的静态资源加速
已发布版本: https://github.com/InvertGeek/mixlink/releases

### 原理
- 用户访问站点时，会转发原站返回内容，相当于一个基础的反向代理服务。  
- 当访问的 URL 命中特定后缀时，会将原站资源上传至加速服务。  
- 下次访问相同资源时，直接重定向到加速地址，不消耗服务器流量。  

### 部署
1. 需要先部署上传服务，推荐使用 [directlink](https://github.com/InvertGeek/directlink)。  
2. 上传服务接口填写为：`directlink服务地址 + /api/upload`。  
3. 运行后编辑 `config.yaml` 文件，自行配置即可。 

### 常见误区
应用是基于反向代理的，不是说部署之后配置文件写上网站域名。原域名网站就直接加速了 \
神仙都做不到这种操作
#### 正常操作
例如您的网址是部署在80端口，默认对外开放 \
先将网站修改到其他端口,例如81 \
然后将mixlink原站地址设置为81端口的原站 \
推荐使用nginx再次反代mixlink,可添加证书和自定义请求头 \
mixlink端口设置为82,然后nginx添加反向代理，监听80端口，代理82端口 \
访问过程: 用户->nginx->mixlink->原站 \
如果原站本身就是nginx也可以这样操作,nginx反代自己都没问题 \
有什么不懂的可以问AI

### 配置文件域名
自定义host规则指的是host请求头匹配某字符串时回源的站点 \
`_default` 代表默认站点,不知道什么是host请求头的可以问AI \
不是随便写个名字，例如默认的www_example_com是指域名为www.example.com时回源的原站


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
