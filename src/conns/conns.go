package ns

import (
	"net"
	"sync"
	"common"
)

type connWrap struct {
	Conn net.Conn
	SendCh chan *common.MsgWrap
	RcvCh chan *common.MsgWrap
	ErrCh chan error
	CloseCh chan bool
}

var conns = map[string]*connWrap{}
var connsL = sync.RWMutex{}

// IsConnected returns whether or not a connection to the given address is
// currently active
func IsConnected(n string) bool {
	connsL.RLock()
	defer connsL.RUnlock()
	_,ok := conns[n]
	return ok
}

// Add connects to the give remote address and adds it to the connection table,
// at the same time setting up a connection spin routing to read and write
// messages on the connection
func Add(raddr string) (bool,error) {
	connsL.Lock()
	defer connsL.Unlock()

	if _,ok := conns[raddr]; ok {
		return true,nil
	}

	c,err := net.Dial("tcp",raddr)
	if err != nil {
		return false,err
	}

	cw := &connWrap{
		Conn: c,
		SendCh: make(chan *common.MsgWrap),
		RcvCh: make(chan *common.MsgWrap),
		ErrCh: make(chan error),
		CloseCh: make(chan bool),
	}
	conns[raddr] = cw

	go connSpin(cw)

	return false,nil

}

// Get gets the connWrap struct for an address, and bool for whether it was
// actually in the table
func Get(raddr string) (*connWrap,bool) {
	connsL.RLock()	
	defer connsL.RUnlock()

	cw,err := conns[raddr]
	return cw,err
}

// Remove removes a connection from the table and closes down the connection
// spin connection
func Remove(raddr string) bool {
	connsL.Lock()
	defer connsL.Unlock()

	if _,ok := conns[raddr]; ok {
		delete(conns,raddr)
		return true
	}

	return false
}
