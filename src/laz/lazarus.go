package laz

import (
	"conns"
	"time"
	"sync"
)

var stopChs = map[string]chan bool{}
var stopLock = sync.RWMutex{}


// AddConn tells lazarus to set up a resurrection loop for the given remote
// address
func AddConn(raddr string) {
	errCh := make(chan error)
	stopCh := make(chan bool)

	stopLock.Lock()
	defer stopLock.Unlock()
	stopChs[raddr] = stopCh

	go connResurectLoop(raddr,errCh,stopCh)
}

// RemoveConn tells lazarus, if it exists, to close the connection to raddr and
// stop resurrecting it
func RemoveConn(raddr string) {
	stopLock.RLock()
	defer stopLock.RUnlock()

	if stopCh,ok := stopChs[raddr]; ok {
		close(stopCh)
	}
}

func connResurectLoop(raddr string, errCh chan error, stopCh chan bool) {
	for {
		go connLoop(raddr, errCh, stopCh)
		select {
			case _,ok := <- stopCh:
				if !ok {
					break
				}
		}
	}
	close(errCh)
}


func connLoop(raddr string, errCh chan error, stopCh chan bool) {

	// Returning means the connection is fucked and we're gonna remake it. Wait
	// both to hackily avoid race conditions and to not spam the remote server
	// if it's having a bad day
	defer conns.Remove(raddr)
	defer time.Sleep(2 * time.Second)
	defer sendStop(stopCh)

	cw,err := conns.Add(raddr)
	if err != nil {
		errCh <- err
		return
	}

	var ok bool
	for {
		select {
			case err,ok = <- cw.ErrCh:
				if ok {
					errCh <- err
				}
			case _,ok = <- stopCh:
		}

		if !ok {
			break
		}
	}
}

func sendStop(stopCh chan bool) {
	stopCh <- true
}
