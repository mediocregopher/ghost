package conns

import (
	"net"
	"encoding/gob"
	"common"
	"io"
)

// Listen takes in an address string and listens on that address for new
// connections, creating the channels that those incoming connections will have
// their messages sent to
func Listen(addr string) (chan *interface{},chan error,error) {

	l,err := net.Listen("tcp",addr)
	if err != nil {
		return nil,nil,err
	}

	rcvCh := make(chan *interface{})
	errCh := make(chan error)
	
	go listenLoop(l,rcvCh,errCh)
	return rcvCh,errCh,nil
}

func listenLoop(l net.Listener, rcvCh chan *interface{}, errCh chan error) {
	for {
		conn,err := l.Accept()
		if err != nil {
			errCh <- err
		} else {
			go connLoop(conn,rcvCh,errCh)
		}
	}
}

func connLoop(conn net.Conn, rcvCh chan *interface{}, errCh chan error) {
	dec := gob.NewDecoder(conn)
	for {
		var msgwrap common.MsgWrap
		err := dec.Decode(&msgwrap)
		if err == io.EOF {
			break
		} else if err != nil {
			errCh <- err
		} else {
			rcvCh <- &msgwrap.Msg
		}
	}
}
