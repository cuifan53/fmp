package fmp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
)

type Conn struct {
	server  *Server
	tcpConn *net.TCPConn
	mn      string
	msgChan chan []byte
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
}

func newConn(server *Server, tcpConn *net.TCPConn) *Conn {
	c := &Conn{
		server:  server,
		tcpConn: tcpConn,
		msgChan: make(chan []byte),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func (c *Conn) start() {
	go c.reader()
	go c.writer()
	go c.server.handler.OnOpened(c)
}

func (c *Conn) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed == true {
		return
	}
	if err := c.tcpConn.Close(); err != nil {
		l.Error(err.Error())
		return
	}
	c.closed = true
	c.cancel()
	// 连接超时情况 旧连接还未断开 新连接已连上
	nowF, err := c.tcpConn.File()
	if err != nil {
		l.Error(err.Error())
		return
	}
	newF, err := c.server.GetConn(c.mn).tcpConn.File()
	if err != nil {
		l.Error(err.Error())
		return
	}
	// 如果server connMap里还是旧连接(没有新连接连上)
	if nowF.Fd() == newF.Fd() {
		c.server.removeConn(c.mn)
		go c.server.handler.OnClosed(c)
	}
}

func (c *Conn) reader() {
	defer c.stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.receiveMsg()
			if err == io.EOF { // 客户端断开连接
				l.Warning("conn " + c.mn + " is stopped")
				return
			}
			if err != nil {
				l.Error(err.Error())
				continue
			}
			if msg == nil {
				continue
			}
			// 触发mn变化回调
			if c.mn == "" {
				c.mn = msg.parsedData.Mn
				// 如果此mn不存在连接 则执行OnMn回调
				if exist := c.server.GetConn(c.mn); exist == nil {
					go c.server.handler.OnMn(c)
				}
				c.server.addConn(c)
			}
			// 触发接收报文回调
			if err := c.server.antsPool.Invoke(msg); err != nil {
				l.Error(err.Error())
				continue
			}
		}
	}
}

func (c *Conn) writer() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.msgChan:
			if _, err := c.tcpConn.Write(data); err != nil {
				l.Error(err.Error())
				return
			}
		}
	}
}

func (c *Conn) receiveMsg() (*Msg, error) {
	// 包头+数据段长度部分
	headData := make([]byte, MsgHeaderLen+MsgDataLenLen)
	if _, err := io.ReadFull(c.tcpConn, headData); err != nil {
		return nil, err
	}
	dataLen, err := strconv.Atoi(string(headData)[MsgHeaderLen:])
	if err != nil {
		return nil, err
	}
	// 数据段+crc段+eof结尾部分
	data := make([]byte, dataLen+MsgCrcLen+MsgEofLen)
	if _, err := io.ReadFull(c.tcpConn, data); err != nil {
		return nil, err
	}
	msg := &Msg{
		conn: c,
		// Data 不包含Eof结尾
		data: bytes.Join([][]byte{headData, data}, []byte{})[:MsgHeaderLen+MsgDataLenLen+dataLen+MsgCrcLen],
	}
	parsedData, err := U.parse(string(msg.data))
	if err != nil {
		return nil, err
	}
	msg.parsedData = parsedData
	return msg, nil
}

func (c *Conn) SendMsg(data string) error {
	c.mu.RLock()
	if c.closed == true {
		c.mu.RUnlock()
		return errors.New("mn " + c.mn + " conn closed when send msg")
	}
	c.mu.RUnlock()
	c.msgChan <- []byte(U.Pack(data))
	return nil
}

func (c *Conn) GetMn() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mn
}

func (c *Conn) RemoteAddr() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tcpConn.RemoteAddr().String()
}
