package fmp

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"

	"github.com/google/uuid"
)

type Conn struct {
	server      *Server
	tcpConn     *net.TCPConn
	connId      string
	mn          string
	recvBuf     []byte
	sendMsgChan chan []byte
	closed      bool
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func newConn(server *Server, tcpConn *net.TCPConn) *Conn {
	c := &Conn{
		server:      server,
		tcpConn:     tcpConn,
		connId:      uuid.New().String(),
		recvBuf:     make([]byte, 8192),
		sendMsgChan: make(chan []byte),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func (c *Conn) start() {
	go c.reader()
	go c.dealRecvBuf()
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
	// 连接超时情况 旧连接还未断开 新连接已连上 如果server connMap里还是旧连接(没有新连接连上)
	if c.mn != "" && c.connId == c.server.GetConn(c.mn).connId {
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
			_, err := c.tcpConn.Read(c.recvBuf)
			if err != nil {
				l.Warning("conn " + c.mn + " is stopped")
				l.Error(err.Error())
				return
			}
		}
	}
}

func (c *Conn) dealRecvBuf() {
	defer c.stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if !bytes.Contains(c.recvBuf, []byte(MsgEof)) {
				continue
			}
			eofIdx := bytes.Index(c.recvBuf, []byte(MsgEof))
			originMsg := c.recvBuf[:eofIdx]
			parsedData, err := U.parse(string(originMsg))
			if err != nil {
				l.Error(err.Error())
				return
			}
			msg := &Msg{
				conn:       c,
				data:       originMsg,
				parsedData: parsedData,
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
			// 去掉已处理报文 并扩容至原始长度
			c.recvBuf = bytes.Join([][]byte{c.recvBuf[eofIdx+MsgEofLen:], make([]byte, eofIdx+MsgEofLen)}, []byte{})
		}
	}
}

func (c *Conn) writer() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.sendMsgChan:
			if _, err := c.tcpConn.Write(data); err != nil {
				l.Error(err.Error())
				return
			}
		}
	}
}

func (c *Conn) SendMsg(data string) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return errors.New("mn " + c.mn + " conn closed when send msg")
	}
	c.mu.RUnlock()
	c.sendMsgChan <- []byte(U.Pack(data))
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
