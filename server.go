package fmp

import (
	"fmt"
	"net"
	"sync"

	"github.com/panjf2000/ants/v2"
)

type IEventHandler interface {
	// 打开连接时调用
	OnOpened(conn *Conn)
	// 关闭连接时调用
	OnClosed(conn *Conn)
	// MN发生变化时调用
	OnMn(conn *Conn)
	// 接收报文时调用
	React(msg *Msg)
}

type Server struct {
	port     int
	connMap  map[string]*Conn
	handler  IEventHandler
	mu       sync.RWMutex
	antsPool *ants.PoolWithFunc
}

func NewServer(port, poolSize int, handler IEventHandler) *Server {
	antsPool, err := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		handler.React(i.(*Msg))
	})
	if err != nil {
		panic(err)
	}
	return &Server{
		port:     port,
		connMap:  make(map[string]*Conn),
		handler:  handler,
		antsPool: antsPool,
	}
}

func (s *Server) Serve() {
	defer s.antsPool.Release()
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		panic(err)
	}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	l.Info(fmt.Sprintf("fmp-server start success. listening port: %d", s.port))
	for {
		tcpConn, err := lis.AcceptTCP()
		if err != nil {
			l.Error(err.Error())
			continue
		}
		conn := newConn(s, tcpConn)
		go conn.start()
	}
}

func (s *Server) addConn(conn *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connMap[conn.mn] = conn
}

func (s *Server) removeConn(mn string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connMap, mn)
}

func (s *Server) GetConn(mn string) *Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connMap[mn]
}
