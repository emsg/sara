package types

const (
	END_FLAG   byte = 1
	HEART_BEAT byte = 2
	KILL       byte = 3

	/*
	   %% 0····打开session
	   %% 1····普通聊天，文本
	   %% 2····群聊，文本
	   %% 3····状态同步
	   %% 4····系统消息
	*/
	MSG_TYPE_OPEN_SESSION uint = 0
	MSG_TYPE_CHAT              = 1
	MSG_TYPE_GROUP_CHAT        = 2
	MSG_TYPE_STATE             = 3
	MSG_TYPE_SYSTEM            = 4

	//session login fail reason
	FAIL_TIMEOUT string = "timeout"     //规定时间内没有发送 “打开会话” 请求
	FAIL_TYPE           = "fail_type"   //第一个请求应该是type=0，否则返回此错误
	FAIL_TOKEN          = "fail_token"  //inner_token过期或失效
	FAIL_PARAM          = "fail_param"  //属性不符合规则
	FAIL_PACKET         = "fail_packet" //数据包与协议不符
	NORMAL              = "normal"

	//session status
	STATUS_CONN  = "conn"
	STATUS_LOGIN = "login"
	STATUS_CLOSE = "close"

	//store perfix
	SESSION_PERFIX = "session_"
	OFFLINE_PERFIX = "offline_"

	// for node/router.go
	ROUTER_QUEUE_SIZE int = 100000
	ROUTER_WORKER_NUM     = 500
	// for node/node.go
	DATACHANNEL_QUEUE_SIZE = 20000
	ACCEPT_QUEUE_SIZE      = 65536
	ACCEPT_WORKER_NUM      = 100
	// for saradb/database.go
	DB_W_QUEUE_SIZE = 20000
)
