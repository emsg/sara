package node

import (
	"github.com/alecthomas/log4go"
	"sara/core"
	"sync"
)

var sessionMap map[string]*core.Session = make(map[string]*core.Session)
var lock *sync.RWMutex = new(sync.RWMutex)

func registerSession(session *core.Session) {
	if sid := session.Status.Sid; sid != "" {
		log4go.Debug("reg_session sid=%s", sid)
		lock.Lock()
		sessionMap[sid] = session
		lock.Unlock()
	}
}

func fetchSession(sid string) (session *core.Session, ok bool) {
	lock.RLock()
	defer lock.RUnlock()
	session, ok = sessionMap[sid]
	return
}
