package conn

import (
	"encoding/gob"
	"errors"
	"github.com/mediocregopher/ghost/common"
	"net"
	"time"
)

type msgWrap struct {
	msg   interface{}
	retCh chan error
}

// A struct that holds all information needed to keep an outgoing ghost
// connection alive
type Conn struct {
	addr    string
	conn    net.Conn
	msgCh   chan *msgWrap
	closeCh chan struct{}
	enc     *gob.Encoder
}

// Returns a new connection struct. This connection may error out on the first
// attempt, in which case it is effectively closed and an error is returned. If
// not, it may close later on in its live, in which case it will be attempted to
// be resurrected
func New(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := Conn{
		addr:    addr,
		conn:    conn,
		msgCh:   make(chan *msgWrap),
		closeCh: make(chan struct{}),
		enc:     gob.NewEncoder(conn),
	}

	go c.spin()
	return &c, nil
}

// spin has the following behavior. It will pull messages off of the msgCh until
// either the connection hits EOF or the closeCh is closed. If clseCh is closed
// then we break out of the whole thing, Close the conn, and die. If the
// connection hits EOF we stop reading from the msgCh and keep trying to create
// a new connection until we're successful, at which point we go back to reading
// off of the msgCh. While we're creating a new connection we check each time if
// the closeCh has been closed; if it has we break out of the whole thing
//
// The following are consequences:
// - If someone is trying to write a message, they may be waiting a while due to
//   resurrection. They should always have some kind of timeout for writing.
// - Another reason they could wait a while is if the connection is closed just
//   they write. Another good reason to have a timeout
func (c *Conn) spin() {
	var m *msgWrap
resloop:
	for {
		for {
			select {
			case m = <-c.msgCh:
			case <-c.closeCh:
				break resloop
			}

			mwrap := common.MsgWrap{m.msg}
			err := c.enc.Encode(mwrap)
			if m.retCh != nil {
				m.retCh <- err
			}
			if err != nil && !c.IsClosed() {
				break
			}
		}

		for !c.IsClosed() {
			conn, err := net.Dial("tcp", c.addr)
			if err == nil {
				c.conn = conn
				c.enc = gob.NewEncoder(conn)
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}

	c.conn.Close()
}

// IsClosed returns whether or not the connection has been closed. This does not
// reflect whether the physical tcp connection is currently closed, because it
// could be in the process of being resurrected.
func (c *Conn) IsClosed() bool {
	select {
	case <-c.closeCh:
		return true
	default:
		return false
	}
}

// Close closes the connection if it's open, and stops resurrection if it isn't.
func (c *Conn) Close() {
	close(c.closeCh)
}

// Send sends an arbitrary message to the connection. It could take up to 10
// seconds to return, and will return any errors that may occur
func (c *Conn) Send(msg interface{}) error {
	m := msgWrap{msg, make(chan error)}
	select {
	case c.msgCh <- &m:
	case <-time.After(5 * time.Second):
		return errors.New("sending message timedout")
	}
	select {
	case err := <-m.retCh:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("receiving error code from message timedout")
	}

	return nil
}

// SendAsync sends an arbitrary message to the connection. It will return almost
// immediately, if the connection is ready for a message the message will be
// dropped.
func (c *Conn) SendAsync(msg interface{}) {
	m := msgWrap{msg, nil}
	select {
	case c.msgCh <- &m:
	case <-time.After(100 * time.Millisecond):
	}
}
