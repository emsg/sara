# sara
<p>由 golang 实现的 emsg 协议 server</p>
<p>原版的 emsg_server 是使用 erlang 编写的云服务平台，有版权问题，无法开源，所以 golang 重构一下为了学习与交流<p> 
<p>任何人都可以无条件使用本服务，并可以为 sara 贡献代码，我会认真审核您提交的代码.</p>
# 协议文档
https://github.com/emsg/docs/wiki

## 安装与使用 (linux/macOs)
#### 环境依赖
```sh
golang 1.7+
redis
```

#### 编译 
###### 假设 GOPATH 在 /opt/gopath 目录, golang 1.7+
```sh
cd /opt/gopath/src
git clone https://github.com/emsg/sara.git
cd /opt/gopath
go install sara
# 应当确保 /usr/local/bin 在 PATH 中
sudo ln -s /usr/local/bin/sara /opt/gopath/bin/sara
```

#### 运行
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
     pprof  将 cpu/mem/block 信息写入文件

GLOBAL OPTIONS:
   --debug                   write 'pprof' info to /tmp/sara_cpu.out and /tmp/sara_mem.out
   --config value, -c value  set config path  (default: "/etc/sara/conf.json")
   --help, -h                show help
   --version, -v             show current version
```

###### sara setup -h
```
NAME:
   sara setup - 生成默认配置文件

USAGE:
   sara setup [command options] [arguments...]

OPTIONS:
   --out value, -o value  配置文件全路径 (default: "/etc/sara/conf.json")
```
###### 配置说明 : /etc/sara/conf.json
<table>
<tr><th>参数</th><th>默认值</th><th>说明</th></tr>
<tr><td>port</td><td>4222</td><td>tcp 服务端口</td></tr>
<tr><td>wsport</td><td>4224</td><td>websocket 服务端口</td></tr>
<tr><td>tlsport</td><td>4333</td><td>tls 服务端口，单向认证</td></tr>
<tr><td>wssport</td><td>4334</td><td>wss 服务端口，与tls使用同一个证书</td></tr>
<tr><td>rpcport</td><td>4280</td><td>https://github.com/emsg/docs/wiki/RPC 功能接口</td></tr>
<tr><td>accesstoken</td><td></td><td>调用RPC接口时提供的身份认证</td></tr>
<tr><td>nodeid</td><td>n01</td><td>节点唯一标示，做集群时必须确保此属性唯一</td></tr>
<tr><td>dbaddr</td><td>localhost:6379</td><td>redis地址，不启用 auth，支持单节点和 cluster </td></tr>
<tr><td>dbpool</td><td>100</td><td>redis连接池大小</td></tr>
<tr><td>callback</td><td></td><td>https://github.com/emsg/docs/wiki/RPC 回调接口</td><td> 
<tr><td>nodeaddr</td><td>localhost:4281</td><td>节点间通信地址，做集群部署时使用</td></tr>
<tr><td>logfile</td><td>/tmp/sara.log</td><td>日志文件</td></tr>
<tr><td>loglevel</td><td>3</td><td>0:ERROR,1:WRAN,2:INFO,3:DEBUG</td></tr>
<tr><td>dc</td><td>dc01</td><td>TODO:数据中心编号，跨数据中心部署</td></tr>
<tr><td>keyfile</td><td>/etc/sara/server.key</td><td>私钥: openssl genrsa -out server.key 2048</td></tr>
<tr><td>certfile</td><td>/etc/sara/server.pem</td><td>证书: openssl req -new -x509 -key server.key -out server.pem -days 3650</td></tr>
<tr><td>enable_tcp</td><td>true</td><td>true:提供tcp服务,false:不提供tcp服务</td></tr>
<tr><td>enable_tls</td><td>false</td><td>true:需要提供 keyfile 和 certfile，false:关闭 tls 服务</td></tr>
<tr><td>enable_ws</td><td>true</td><td>true:提供websocket服务，false:不提供ws服务</td></tr>
<tr><td>enable_wss</td><td>false</td><td>true:需要提供 keyfile 和 certfile，false:关闭 tls 服务</td></tr>
<tr><td>enable_auth</td><td>false</td><td>开启认证，需要提供 callback 参数，并实现 auth 接口</td></tr>
<tr><td>enable_offline_callback</td><td>false</td><td>开启离线消息回调，需要提供 callback 参数，并实现 offline 接口</td></tr>
</table>

###### 启动服务
```sh

#> sara 
[16:19:58 CST 2017/02/10] [INFO] (sara/saradb.(*SaraDatabase).wbfConsumer:110) write buffer started ; total consume [40]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).cleanGhostSession:353) register node : n01
[16:19:58 CST 2017/02/10] [INFO] (sara/sararpc.(*RPCServer).Start:47) RPCServer listener on  [localhost:4281]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).cleanGhostSession:355) 🔪  👻  clean ghost session
[16:19:58 CST 2017/02/10] [INFO] (sara/service.StartRPC:28) http-rpc start on [0.0.0.0:4280]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).StartTCP:79) tcp start on [0.0.0.0:4222]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).StartWS:70) ws start on [4224]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).StartTLS:136) tls start on [0.0.0.0:4333]
[16:19:58 CST 2017/02/10] [INFO] (sara/node.(*Node).StartWSS:113) wss start on [0.0.0.0:4334]

```
###### 在后台运行： nohup sara > /tmp/sara.log &

#### 集群
###### 有关集群的配置项
```sh
accesstoken : 没个节点的 token 都应当一致，否则节点间也无法通信;
nodeid : 集群中每个节点都有一个唯一的 id ，切记不能重复，建议按照 n01、n02、n03 这样编排;
dbaddr : 每个节点都要把 session 注册到这个 db 中，所以每个节点的此项配置应当是一致的;
```

