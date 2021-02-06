package fmp

import (
	"fmt"
	"net"
	"sync"
)

type IEventHandler interface {
	// 打开连接时调用
	OnOpened(conn *Conn)
	// 关闭连接时调用
	OnClosed(conn *Conn)
	// MN发生变化时调用
	OnMn(conn *Conn)
	// 接收报文时调用
	React(conn *Conn, msg *Msg)
}

type Server struct {
	port    int
	connMap map[string]*Conn
	handler IEventHandler
	mu      sync.RWMutex
}

func NewServer(port int, handler IEventHandler) *Server {
	return &Server{
		port:    port,
		connMap: make(map[string]*Conn),
		handler: handler,
	}
}

func (s *Server) Serve() {
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
