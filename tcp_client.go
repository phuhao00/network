package network

import (
	"github.com/phuhao00/spoor"
	"github.com/phuhao00/spoor/logger"
	"net"
	"runtime/debug"
	"sync/atomic"
)

type TcpClient struct {
	*TcpSession
	Address         string
	ChMsg           chan *Message
	OnMessageCb     func(message *Packet)
	logger          *spoor.Spoor
	bufferSize      int
	running         atomic.Value
	OnCloseCallBack func()
	closed          int32
}

func NewClient(address string, connBuffSize int, logger *spoor.Spoor) *TcpClient {
	client := &TcpClient{
		bufferSize: connBuffSize,
		Address:    address,
		logger:     logger,
		TcpSession: nil,
	}
	client.running.Store(false)
	return client
}

func (c *TcpClient) Dial() (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", c.Address)

	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp6", nil, tcpAddr)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *TcpClient) Run() {
	conn, err := c.Dial()
	if err != nil {
		logger.Error("%v", err)
		return
	}
	tcpSession, err := NewTcpSession(conn, c.bufferSize, c.logger)
	if err != nil {
		logger.Error("%v", err)
		return
	}
	c.TcpSession = tcpSession
	c.Impl = c
	c.Reset()
	c.running.Store(true)
	go c.Connect()
}

func (c *TcpClient) OnClose() {
	if c.OnCloseCallBack != nil {
		c.OnCloseCallBack()
	}
	c.running.Store(false)
	c.TcpSession.OnClose()
}

func (c *TcpClient) OnMessage(data *Message, conn *TcpSession) {

	c.Verify()

	defer func() {
		if err := recover(); err != nil {
			logger.Error("[OnMessage] panic ", err, "\n", string(debug.Stack()))
		}
	}()

	c.OnMessageCb(&Packet{
		Msg:  data,
		Conn: conn,
	})
}

// Close 关闭连接
func (c *TcpClient) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		err := c.Conn.Close()
		if err != nil {
			logger.Error("")
		}
		close(c.stopped)
	}
}

func (c *TcpClient) IsRunning() bool {
	return c.running.Load().(bool)
}
