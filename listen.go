package ghost

import (
	"encoding/gob"
	"github.com/mediocregopher/ghost/common"
	"io"
	"net"
)

// Listen for ghost connections on a given address.
//
// Returns:
//
// chan *interface{}: channel where messages sent by the connected ghost
// connections get sent to. Make sure all messages types that you expect are
// Register()'d.
//
// chan error: channel where error messages (apart from connection closed
// messages) are sent to. At the point that you'll be getting them there's not
// much you can do, but you have access in case you want to log them or
// something.
//
// error: In case there's an error opening up the listen socket
//
// It's important that the message and error channels are always read from,
// otherwise it could (read: will!) block up internal processes
func Listen(addr string) (chan interface{}, chan error, error) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	rcvCh := make(chan interface{})
	errCh := make(chan error)

	go listenLoop(l, rcvCh, errCh)
	return rcvCh, errCh, nil
}

func listenLoop(l net.Listener, rcvCh chan interface{}, errCh chan error) {
	for {
		conn, err := l.Accept()
		if err != nil {
			errCh <- err
		} else {
			go connLoop(conn, rcvCh, errCh)
		}
	}
}

func connLoop(conn net.Conn, rcvCh chan interface{}, errCh chan error) {
	dec := gob.NewDecoder(conn)
	for {
		var msgwrap common.MsgWrap
		err := dec.Decode(&msgwrap)
		if err == io.EOF {
			break
		} else if err != nil {
			errCh <- err
		} else {
			rcvCh <- msgwrap.Msg
		}
	}
}
