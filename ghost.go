package ghost

import (
	"github.com/mediocregopher/ghost/src/conns"
	"github.com/mediocregopher/ghost/src/laz"
)

// Listen for ghost connections on a given address.
//
// Returns:
//
// chan *interface{}
//		channel where messages sent by the connected ghost connections get sent
//		to. Make sure all messages types that you expect are Register()'d.
//
// chan error
//		channel where error messages (apart from connection closed messages) are
//		sent to. At the point that you'll be getting them there's not much you
//		can do, but you have access in case you want to log them or something.
//
// error
//		In case there's an error opening up the listen socket
//
// It's important that the message and error channels are always read from,
// otherwise it could (read: will!) block up internal processes
func Listen(addr string) (chan interface{}, chan error, error) {
	return conns.Listen(addr)
}

// AddConn creates a connection to the given remote address that will be
// automatically resurrected if it gets disconnected. Once you've called this
// you can call Send on the address and send it (almost) arbitrary data, as long
// as you've Register()'d those structures on the other end.
//
// NOTE: This call is asynchronous, so the connection may not yet be established
// if you try Send immediately after calling it
func AddConn(raddr string) {
	laz.AddConn(raddr)
}

// RemoveConn closes and removes the connection to the given remote address,
// assuming it was opened by AddConn previously
func RemoveConn(raddr string) {
	laz.RemoveConn(raddr)
}

// Send sends an arbitrary message to the given remote location, assuming
// AddConn has been called on it already. Returns any errors the socket might
// give when trying to send the message. io.EOF could be one of the errors, if
// the socket closes, but in that case it's not necessary to call AddConn again.
func Send(raddr string, msg interface{}) error {
	return conns.Send(raddr, msg)
}

// SendAll sendss a message to all addresses that have had AddConn called on
// them so far.  It doesn't handle errors in any way, so if you want to do that
// you'll have to keep your own list of addresses and loop through that
// manually.
func SendAll(msg interface{}) {
	conns.SendAll(msg)
}

// Register takes in an instance of a type and registers that type with the
// decoder as a potential candidate for decryption. This must be done on all
// types that will be received on the *interface{} channel returned by listen.
func Register(something interface{}) {
	conns.Register(something)
}
