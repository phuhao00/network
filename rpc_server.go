package network

import (
	"github.com/phuhao00/spoor/logger"
	"net"
	"net/rpc"
	"runtime/debug"
	"sync"
	"time"
)

type RpcServer struct {
	Addr string
	ln   *net.TCPListener
	wgLn sync.WaitGroup
}

func (srv *RpcServer) Init(addr string) {

	srv.Addr = addr

	tcpAddr, err := net.ResolveTCPAddr("tcp4", srv.Addr)

	if err != nil {
		logger.Error("[net] addr resolve error", tcpAddr, err)
		return
	}

	ln, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		logger.Error("%v", err)
		return
	}

	srv.ln = ln
	logger.Info("RpcServer Listen %s", srv.ln.Addr().String())
}

func (srv *RpcServer) Run() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[net] panic", err, "\n", string(debug.Stack()))
		}
	}()

	srv.wgLn.Add(1)
	defer srv.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := srv.ln.AcceptTCP()

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logger.Warn("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		err = conn.SetKeepAlive(true)
		if err != nil {
			logger.Error(err.Error())
		}

		err = conn.SetKeepAlivePeriod(1 * time.Minute)
		if err != nil {
			logger.Error(err.Error())
		}
		err = conn.SetNoDelay(true)
		if err != nil {
			logger.Error(err.Error())
		}

		err = conn.SetWriteBuffer(128 * 1024)
		if err != nil {
			logger.Error(err.Error())
		}

		err = conn.SetReadBuffer(128 * 1024)
		if err != nil {
			logger.Error(err.Error())
		}

		go rpc.ServeConn(conn)

		logger.Debug("accept a rpc conn")
	}
}

func (srv *RpcServer) Close() {
	err := srv.ln.Close()
	if err != nil {
		logger.Error(err.Error())
	}
	srv.wgLn.Wait()
}
