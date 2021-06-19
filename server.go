package fmp

import (
	"errors"
	"sync"

	"github.com/cuifan53/fmp/protocol"

	"github.com/panjf2000/gnet"
)

func NewServer(port string, protocol protocol.IProtocol, handler IEventHandler) *Server {
	if protocol == nil {
		panic("protocol incorrect")
	}
	if handler == nil {
		panic("handler incorrect")
	}
	return &Server{
		connMap:  make(map[string]gnet.Conn),
		port:     port,
		protocol: protocol,
		handler:  handler,
	}
}

type Server struct {
	*gnet.EventServer
	mu       sync.RWMutex
	connMap  map[string]gnet.Conn
	port     string
	protocol protocol.IProtocol
	handler  IEventHandler
}

func (s *Server) Serve() {
	if err := gnet.Serve(
		s,
		s.port,
		gnet.WithCodec(&delimiterCodec{delimiter: s.protocol.Eof()}),
		gnet.WithReusePort(true),
	); err != nil {
		panic(err)
	}
}

// Send 发送报文
func (s *Server) Send(mn, data string) error {
	conn := s.GetConn(mn)
	if conn == nil {
		return errors.New("mn: " + mn + ", connection incorrect")
	}
	return conn.AsyncWrite(s.protocol.Pack(data))
}

// GetMns 获取所有MN
func (s *Server) GetMns() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mns := make([]string, 0)
	for mn, _ := range s.connMap {
		if mn != "" {
			mns = append(mns, mn)
		}
	}
	return mns
}

// GetConn 通过MN获取Tcp连接
func (s *Server) GetConn(mn string) gnet.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connMap[mn]
}

func (s *Server) setConn(mn string, conn gnet.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connMap[mn] = conn
}

func (s *Server) removeConn(mn string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connMap, mn)
}

// ** 以下为重写server内部gnet.EventServer方法 ** //

func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	s.handler.OnOpened(c)
	return
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	for mn, conn := range s.connMap {
		if conn == c {
			s.removeConn(mn)
			s.handler.OnMn(mn, false)
		}
	}
	s.handler.OnClosed(c)
	return
}

func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	parsedData, err := s.protocol.Parse(string(frame))
	if err != nil {
		return
	}

	mn := ""
	msg := &Msg{
		data: frame,
	}

	switch s.protocol.Name() {
	case protocol.ProtocolNameNs:
		parsedDataNs, ok := parsedData.(*protocol.ParsedDataNs)
		if !ok {
			return
		}
		msg.parsedDataNs = parsedDataNs
		mn = parsedDataNs.Mn
	case protocol.ProtocolNameRdd:
		parsedDataRdd, ok := parsedData.(*protocol.ParsedDataRdd)
		if !ok {
			return
		}
		msg.parsedDataRdd = parsedDataRdd
		mn = parsedDataRdd.Mn
	default:
		return
	}

	if mn == "" {
		return
	}

	if s.GetConn(mn) != c {
		s.setConn(mn, c)
		s.handler.OnMn(mn, true)
	}
	s.handler.React(c, msg)
	return
}
