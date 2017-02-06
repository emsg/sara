# sara
<p>由 golang 实现的 emsg 协议 server</p>
<p>原版的 emsg_server 是使用 erlang 编写的云服务平台，有版权问题，无法开源，所以 golang 重构一下为了学习与交流<p> 
<p>任何人都可以无条件使用本服务，并可以为 sara 贡献代码，我会认真审核您提交的代码.</p>
# 协议文档
https://github.com/emsg/docs/wiki

## 编译安装

#### 假设 GOPATH 在 /app/gopath 目录
```sh
cd /app/gopath/src
git clone https://github.com/emsg/sara.git
cd /app/gopath
go install sara
ln -s /usr/local/bin/sara /app/gopath/bin/sara
```
