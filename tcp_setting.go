package network

import (
	"net"
	"time"
)

type TCPKeepAliveListener struct {
	*net.TCPListener
}

// Accept accepts the next incoming call and returns the new
// connection. KeepAlivePeriod is set properly.
func (ln TCPKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	err = tc.SetKeepAlive(true)
	if err != nil {
		return
	}
	err = tc.SetKeepAlivePeriod(3 * time.Minute)
	if err != nil {
		return
	}
	return tc, nil
}
