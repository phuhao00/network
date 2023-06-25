package network

import (
	"encoding/binary"
	"fmt"
	"github.com/phuhao00/spoor"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/phuhao00/spoor/logger"
)

const timeoutTime = 30 // 连接通过验证的超时时间

type TcpSession struct {
	Conn        net.Conn
	Impl        IService
	ConnID      int64
	verify      int32
	closed      int32
	stopped     chan bool
	signal      chan interface{}
	lastSignal  chan interface{}
	wgRW        sync.WaitGroup
	msgParser   *BufferPacker
	msgBuffSize int
}

func NewTcpSession(conn *net.TCPConn, msgBuffSize int, logger *spoor.Spoor) (*TcpSession, error) {
	tcpConn := &TcpSession{
		closed:      -1,
		verify:      0,
		stopped:     make(chan bool, 1),
		signal:      make(chan interface{}, 100),
		lastSignal:  make(chan interface{}, 1),
		Conn:        conn,
		msgBuffSize: msgBuffSize,
		msgParser:   newInActionPacker(),
	}
	// Try to open keepalive for tcp.
	err := conn.SetKeepAlive(true)
	if err != nil {
		return nil, err
	}
	err = conn.SetKeepAlivePeriod(1 * time.Minute)
	if err != nil {
		return nil, err
	}
	// disable Nagle algorithm.
	err = conn.SetNoDelay(true)
	if err != nil {
		return nil, err
	}
	err = conn.SetWriteBuffer(msgBuffSize)
	if err != nil {
		return nil, err
	}
	err = conn.SetReadBuffer(msgBuffSize)
	if err != nil {
		return nil, err
	}
	return tcpConn, nil
}

func (c *TcpSession) Connect() {
	if atomic.CompareAndSwapInt32(&c.closed, -1, 0) {
		c.wgRW.Add(1)
		go c.HandleRead()
		c.wgRW.Add(1)
		go c.HandleWrite()
	}
	timeout := time.NewTimer(time.Second * timeoutTime)
L:
	for {
		select {
		// 等待通到返回 返回后检查连接是否验证完成 如果没有验证 则关闭连接
		case <-timeout.C:
			if !c.Verified() {
				logger.Error("[Connect] 验证超时 ip addr %s", c.RemoteAddr())
				c.Close()
				break L
			}
		case <-c.stopped:
			break L
		}
	}
	timeout.Stop()
	c.wgRW.Wait()
	c.Impl.OnClose()
}

func (c *TcpSession) HandleRead() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[HandleRead] panic ", err, "\n", string(debug.Stack()))
		}
	}()
	defer c.Close()

	defer c.wgRW.Done()

	for {
		data, err := c.msgParser.Read(c)
		if err != nil {
			if err != io.EOF {
				logger.Error("read message error: %v", err)
			}
			break
		}
		message, err := c.msgParser.Unpack(data)
		c.Impl.OnMessage(message, c)
	}
}

func (c *TcpSession) HandleWrite() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[HandleWrite] panic", err, "\n", string(debug.Stack()))
		}
	}()
	defer c.Close()
	defer c.wgRW.Done()
	for {
		select {
		case signal := <-c.signal: // 普通消息
			data, ok := signal.([]byte)
			if !ok {
				logger.Error("write message %v error: msg is not bytes", reflect.TypeOf(signal))
				return
			}
			err := c.msgParser.Write(c, data...)
			if err != nil {
				logger.Error("write message %v error: %v", reflect.TypeOf(signal), err)
				return
			}
		case signal := <-c.lastSignal: // 最后一个通知消息
			data, ok := signal.([]byte)
			if !ok {
				logger.Error("write message %v error: msg is not bytes", reflect.TypeOf(signal))
				return
			}
			err := c.msgParser.Write(c, data...)
			if err != nil {
				logger.Error("write message %v error: %v", reflect.TypeOf(signal), err)
				return
			}
			time.Sleep(2 * time.Second)
			return
		case <-c.stopped: // 连接关闭通知
			return
		}
	}
}

func (c *TcpSession) AsyncSend(msgID uint64, msg interface{}) bool {

	if c.IsShutdown() {
		return false
	}

	data, err := c.msgParser.Pack(msgID, msg)
	if err != nil {
		logger.Error("[AsyncSend] Pack msgID:%v and msg to bytes error:%v", msgID, err)
		return false
	}

	if uint32(len(data)) > c.msgParser.maxMsgLen {
		logger.Error("[AsyncSend] 发送的消息包体过长 msgID:%v", msgID)
		return false
	}

	err = c.Signal(data)
	if err != nil {
		c.Close()
		logger.Error("%v", err)
		return false
	}

	return true
}

func (c *TcpSession) AsyncSendRowMsg(data []byte) bool {

	if c.IsShutdown() {
		return false
	}

	if uint32(len(data)) > c.msgParser.maxMsgLen {
		logger.Error("[AsyncSendRowMsg] 发送的消息包体过长 AsyncSendRowMsg")
		return false
	}

	err := c.Signal(data)
	if err != nil {
		c.Close()
		logger.Error("%v", err)
		return false
	}

	return true
}

// AsyncSendLastPacket 缓存在发送队列里等待发送goroutine取出 (发送最后一个消息 发送会关闭tcp连接 终止tcp goroutine)
func (c *TcpSession) AsyncSendLastPacket(msgID uint64, msg interface{}) bool {
	data, err := c.msgParser.Pack(msgID, msg)
	if err != nil {
		logger.Error("[AsyncSendLastPacket] Pack msgID:%v and msg to bytes error:%v", msgID, err)
		return false
	}

	if uint32(len(data)) > c.msgParser.maxMsgLen {
		logger.Error("[AsyncSendLastPacket] 发送的消息包体过长 msgID:%v", msgID)
		return false
	}

	err = c.LastSignal(data)
	if err != nil {
		c.Close()
		logger.Error("%v", err)
		return false
	}

	return true
}

func (c *TcpSession) Signal(signal []byte) error {
	select {
	case c.signal <- signal:
		return nil
	default:
		{
			cmd := binary.LittleEndian.Uint16(signal[2:4])
			return fmt.Errorf("[Signal] buffer full blocking connID:%v cmd:%v", c.ConnID, cmd)
		}
	}
}

func (c *TcpSession) LastSignal(signal []byte) error {
	select {
	case c.lastSignal <- signal:
		return nil
	default:
		{
			cmd := binary.LittleEndian.Uint16(signal[2:4])
			return fmt.Errorf("[LastSignal] buffer full blocking connID:%v cmd:%v", c.ConnID, cmd)
		}
	}
}

func (c *TcpSession) Verified() bool {
	return atomic.LoadInt32(&c.verify) != 0
}

func (c *TcpSession) Verify() {
	atomic.CompareAndSwapInt32(&c.verify, 0, 1)
}

func (c *TcpSession) IsClosed() bool {
	return atomic.LoadInt32(&c.closed) != 0
}

func (c *TcpSession) IsShutdown() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *TcpSession) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.Conn.Close()
		close(c.stopped)
	}
}

func (c *TcpSession) Read(b []byte) (int, error) {
	return c.Conn.Read(b)
}

func (c *TcpSession) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c *TcpSession) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *TcpSession) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *TcpSession) Reset() {
	if atomic.LoadInt32(&c.closed) == -1 {
		return
	}
	c.closed = -1
	c.verify = 0
	c.stopped = make(chan bool, 1)
	c.signal = make(chan interface{}, c.msgBuffSize)
	c.lastSignal = make(chan interface{}, 1)
	c.msgParser.reset()
}

// OnConnect ...
func (c *TcpSession) OnConnect() {
	logger.Error("[OnConnect] 建立连接 local:%s remote:%s", c.LocalAddr(), c.RemoteAddr())
}

func (c *TcpSession) OnClose() {
	logger.Error("[OnConnect] 断开连接 local:%s remote:%s", c.LocalAddr(), c.RemoteAddr())
}
