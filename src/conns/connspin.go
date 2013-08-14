package ns

import (
	"common"
	"encoding/gob"
	"io"
)

type decret struct {
	msg *common.MsgWrap
	err error
}

func connSpin(cw *connWrap) {
	out := gob.NewEncoder(cw.Conn)
	in  := gob.NewDecoder(cw.Conn)
	needRead := true
	decCh := make(chan decret)

	for {
		if needRead {
			go func(){
				var msg *common.MsgWrap
				err := in.Decode(msg)
				decCh <- decret{msg,err}
			}()
			needRead = false
		}

		select {
			case dret := <- decCh:
				if dret.err == io.EOF {
					break
				} else if dret.err != nil {
					cw.ErrCh <- dret.err
				} else {
					cw.RcvCh <- dret.msg
				}
				needRead = true

			case msg := <- cw.SendCh:
				err := out.Encode(msg)
				if err == io.EOF {
					break
				} else if err != nil {
					cw.ErrCh <- err
				}

			case <- cw.CloseCh:
				break
		}
		
	}

	close(cw.RcvCh)
	close(cw.ErrCh)
	//Don't close SendCh in case a thread is about to send on this connection
	//Don't close CloseCh in case multiple threads try to close at once
	cw.Conn.Close()
}


