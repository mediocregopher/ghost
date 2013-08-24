package laz

import (
	"github.com/mediocregopher/ghost/src/conns"
	"time"
	"sync"
)

var stopChs = map[string]chan bool{}
var stopLock = sync.RWMutex{}


// AddConn tells lazarus to set up a resurrection loop for the given remote
// address
func AddConn(raddr string) {
	stopCh := make(chan bool)

	stopLock.Lock()
	defer stopLock.Unlock()
	stopChs[raddr] = stopCh

	go connResurectLoop(raddr,stopCh)
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

func connResurectLoop(raddr string, stopCh chan bool) {
	for {
		go connLoop(raddr, stopCh)
		select {
			case _,ok := <- stopCh:
				if !ok {
					break
				}
		}
	}
}


func connLoop(raddr string, stopCh chan bool) {

	// Returning means the connection is fucked and we're gonna remake it. Wait
	// both to hackily avoid race conditions and to not spam the remote server
	// if it's having a bad day (remember defer statements execute in reverse
	// order that they're defined)
	defer sendStop(stopCh)
	defer time.Sleep(2 * time.Second)
	defer conns.Remove(raddr)

	cw,err := conns.Add(raddr)
	if err != nil {
		return
	}

	var ok bool
	for {
		select {
			case _,ok = <- cw.CloseCh:
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
