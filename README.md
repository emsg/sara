# sara
<p>由 golang 实现的 emsg 协议 server</p>
<p>原版的 emsg_server 是使用 erlang 编写的云服务平台，有版权问题，无法开源，所以 golang 重构一下为了学习与交流<p> 
<p>任何人都可以无条件使用本服务，并可以为 sara 贡献代码，我会认真审核您提交的代码.</p>
# 协议文档
https://github.com/emsg/docs/wiki

## 安装与使用

#### 编译
###### 假设 GOPATH 在 /opt/gopath 目录
```sh
cd /opt/gopath/src
git clone https://github.com/emsg/sara.git
cd /app/gopath
go install sara
# 应当确保 /usr/local/bin 在 PATH 中
sudo ln -s /usr/local/bin/sara /app/gopath/bin/sara
```
###### sara -h
```sh
NAME:
   sara - SARA IM Server

USAGE:
   sara [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR(S):
   liangc <cc14514@icloud.com>

COMMANDS:
     version
     stop     停止服务，尽量避免直接 kill 服务
     setup    生成默认配置文件
     help, h  Shows a list of commands or help for one command

   benchmark:
     makeconn  创建指定个数的连接，测试最大连接数

   debug:
     pprof  将 cpu/mem 信息写入文件

GLOBAL OPTIONS:
   --debug                   write 'pprof' info to /tmp/sara_cpu.out and /tmp/sara_mem.out
   --config value, -c value  cmd-line first, config second  (default: "/etc/sara/conf.json")
   --help, -h                show help
   --version, -v             print the version
```

###### sara setup -h
```
bogon:sara liangc$ sara setup -h
NAME:
   sara setup - 生成默认配置文件

USAGE:
   sara setup [command options] [arguments...]

OPTIONS:
   --out value, -o value  配置文件全路径 (default: "/etc/sara/conf.json")
```
