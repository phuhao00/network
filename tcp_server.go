package network

import (
	"github.com/phuhao00/spoor"
	"github.com/phuhao00/spoor/logger"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type TcpServer struct {
	pid            int64
	Addr           string
	MaxConnNum     int
	ln             *net.TCPListener
	connSet        map[net.Conn]interface{}
	counter        int64
	idCounter      int64
	mutexConn      sync.Mutex
	wgLn           sync.WaitGroup
	wgConn         sync.WaitGroup
	connBuffSize   int
	logger         *spoor.Spoor
	MessageHandler func(packet *Packet)
}

func NewTcpServer(addr string, maxConnNum int, buffSize int, logger *spoor.Spoor) *TcpServer {
	s := &TcpServer{
		Addr:         addr,
		MaxConnNum:   maxConnNum,
		connBuffSize: buffSize,
		logger:       logger,
	}
	s.Init()
	return s
}

func (s *TcpServer) Init() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", s.Addr)

	if err != nil {
		logger.Fatal("[Init] addr resolve error", tcpAddr, err)
	}

	ln, err := net.ListenTCP("tcp6", tcpAddr)

	if err != nil {
		logger.Fatal("%v", err)
	}

	if s.MaxConnNum <= 0 {
		s.MaxConnNum = 100
		logger.Info("[Init] invalid MaxConnNum, reset to %v", s.MaxConnNum)
	}

	s.ln = ln
	s.connSet = make(map[net.Conn]interface{})
	s.counter = 1
	s.idCounter = 1
	s.pid = int64(os.Getpid())
	logger.Info("[Init] TcpServer Listen %s", s.ln.Addr().String())
}

func (s *TcpServer) Run() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[Run] panic", err, "\n", string(debug.Stack()))
		}
	}()

	s.wgLn.Add(1)
	defer s.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := s.ln.AcceptTCP()

		if err != nil {
			if _, ok := err.(net.Error); ok {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logger.Info("[Run]accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		if atomic.LoadInt64(&s.counter) >= int64(s.MaxConnNum) {
			err = conn.Close()
			if err != nil {
				logger.Error("[Run] err:%v", err.Error())
			}
			logger.Info("[Run] too many connections %v", atomic.LoadInt64(&s.counter))
			continue
		}
		tcpConnX, err := NewTcpSession(conn, s.connBuffSize, s.logger)
		if err != nil {
			logger.Error("[Run] err:%v", err)
			return
		}
		s.addConn(conn, tcpConnX)
		tcpConnX.Impl = s
		s.wgConn.Add(1)
		go func() {
			tcpConnX.Connect()
			s.removeConn(conn, tcpConnX)
			s.wgConn.Done()
		}()
	}
}

func (s *TcpServer) Close() {
	err := s.ln.Close()
	if err != nil {
		logger.Error(err.Error())
	}
	s.wgLn.Wait()

	s.mutexConn.Lock()
	for conn := range s.connSet {
		err = conn.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	s.connSet = nil
	s.mutexConn.Unlock()
	s.wgConn.Wait()
}

func (s *TcpServer) addConn(conn net.Conn, tcpSession *TcpSession) {
	s.mutexConn.Lock()
	atomic.AddInt64(&s.counter, 1)
	s.connSet[conn] = conn
	nowTime := time.Now().Unix()
	idCounter := atomic.AddInt64(&s.idCounter, 1)
	connId := (nowTime << 32) | (s.pid << 24) | idCounter
	tcpSession.ConnID = connId
	s.mutexConn.Unlock()
	tcpSession.OnConnect()
}

func (s *TcpServer) removeConn(conn net.Conn, tcpConn *TcpSession) {
	tcpConn.Close()
	s.mutexConn.Lock()
	atomic.AddInt64(&s.counter, -1)
	delete(s.connSet, conn)
	s.mutexConn.Unlock()
}

func (s *TcpServer) OnMessage(message *Message, conn *TcpSession) {
	s.MessageHandler(&Packet{
		Msg:  message,
		Conn: conn,
	})
}

func (s *TcpServer) OnClose() {

}

func (s *TcpServer) OnConnect() {

}
