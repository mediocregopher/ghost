package ghost

import (
	"encoding/gob"
	"errors"
	"github.com/mediocregopher/ghost/conn"
	"sync"
)

var conns = map[string]*conn.Conn{}
var connsL = sync.RWMutex{}

// AddConn creates a resurrecting conn and adds it to the pool of connected
// Conns
func AddConn(addr *string) error {
	connsL.Lock()
	defer connsL.Unlock()

	if _, ok := conns[*addr]; ok {
		return nil
	}

	conn, err := conn.New(*addr)
	if err != nil {
		return err
	}

	conns[*addr] = conn
	return nil
}

// RemConn stops a connection, assuming it's in the pool, so that it will be
// closed and resurrection on it will stop
func RemConn(addr *string) {
	connsL.Lock()
	defer connsL.Unlock()

	if conn, ok := conns[*addr]; ok {
		conn.Close()
	}
}

// Send retrieves a connection from the pool and sends a message to it,
// returning any errors from it. If no errors are returned it is safe to assume
// the message was sent successfully
func Send(addr *string, msg interface{}) error {
	connsL.RLock()
	defer connsL.RUnlock()

	if conn, ok := conns[*addr]; ok {
		return conn.Send(msg)
	} else {
		return errors.New("addr not in pool")
	}
}

// SendAll sends a message (asynchronously) to all connections currently in the
// pool
func SendAll(msg interface{}) {
	connsL.RLock()
	defer connsL.RUnlock()

	for _, conn := range conns {
		conn.SendAsync(msg)
	}
}

// Register takes in an instance of a type and registers that type with the
// decoder as a potential candidate for decoding. This must be done on all types
// that will be received on the *interface{} channel returned by listen.
func Register(something interface{}) {
	gob.Register(something)
}
