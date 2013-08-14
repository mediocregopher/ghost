package conns

import (
	"net"
	"sync"
	"common"
	"encoding/gob"
	"io"
	"errors"
)

type connWrap struct {
	Raddr string
	conn net.Conn
	ErrCh chan error
	enc *gob.Encoder
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
// at the same time setting up a connection spin to read errors on the
// connection. Returns the connection wrap object assuming there were no errors
func Add(raddr string) (*connWrap,error) {
	connsL.Lock()
	defer connsL.Unlock()

	if cw,ok := conns[raddr]; ok {
		return cw,nil
	}

	c,err := net.Dial("tcp",raddr)
	if err != nil {
		return nil,err
	}

	cw := &connWrap{
		Raddr: raddr,
		conn: c,
		ErrCh: make(chan error),
		enc: gob.NewEncoder(c),
	}
	conns[raddr] = cw

	go connReadSpin(cw)

	return cw,nil

}

func connReadSpin(cw *connWrap) {
	for {
		// We'll never actually get any data here, but you gotta block on
		// something right?
		b := make([]byte,1)
		_,err := cw.conn.Read(b)

		if err == io.EOF {
			break
		}
	}

	Remove(cw.Raddr)
	close(cw.ErrCh)
	cw.conn.Close()
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

	if cw,ok := conns[raddr]; ok {
		//TODO if someone Removes then immediately Adds a connection it could
		//possibly get Remove'd again by connReadSpin
		cw.conn.Close()
		delete(conns,raddr)
		return true
	}

	return false
}

// Send sends a message to the remote location if the channel is in the table
// and currently available for sending
func Send(raddr string, msg interface{}) error {
	connsL.RLock()
	defer connsL.RUnlock()

	if cw,ok := conns[raddr]; ok {
		return sendDirect(cw,msg)
	} else {
		return errors.New("connection not established")
	}
}

func sendDirect(cw *connWrap, msg interface{}) error {
	msgwrap := &common.MsgWrap{ msg }
	return cw.enc.Encode(msgwrap)
}

// SendAll loops through and asynchronously sends a message to all currently
// active connections
func SendAll(msg interface{}) {
	connsL.RLock()
	defer connsL.RUnlock()

	for _,cw := range conns {
		go sendDirect(cw,msg)
	}
}

// Register registers a type as a potential decode unmarshalling type. This must
// be done for all message types that will be sent.
func Register(something interface{}) {
	gob.Register(something)
}
