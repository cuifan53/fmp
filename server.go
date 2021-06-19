package fmp

import (
	"errors"
	"sync"

	"github.com/cuifan53/fmp/protocol"

	"github.com/panjf2000/gnet"
)

type ProtocolName string

const (
	ProtocolNs  ProtocolName = "Ns"
	ProtocolRdd ProtocolName = "Rdd"
)

func NewServer(port string, protocolName ProtocolName, handler IEventHandler) *Server {
	if handler == nil {
		panic("handler incorrect")
	}
	s := &Server{
		connMap:  make(map[string]gnet.Conn),
		port:     port,
		protocol: protocolName,
		handler:  handler,
	}
	switch protocolName {
	case ProtocolNs:
		s.protocolNs = protocol.NewProtocolNs()
	case ProtocolRdd:
		s.protocolRdd = protocol.NewProtocolRdd()
	default:
		panic("protocol incorrect")
	}
	return s
}

type Server struct {
	*gnet.EventServer
	mu          sync.RWMutex
	connMap     map[string]gnet.Conn
	port        string
	protocol    ProtocolName
	protocolNs  *protocol.Ns
	protocolRdd *protocol.Rdd
	handler     IEventHandler
}

func (s *Server) Serve() {
	var delimiter []byte
	switch s.protocol {
	case ProtocolNs:
		delimiter = s.protocolNs.Eof()
	case ProtocolRdd:
		delimiter = s.protocolRdd.Eof()
	}
	if err := gnet.Serve(
		s,
		s.port,
		gnet.WithCodec(&delimiterCodec{delimiter: delimiter}),
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
	var packed []byte
	switch s.protocol {
	case ProtocolNs:
		packed = s.protocolNs.Pack(data)
	case ProtocolRdd:
		packed = s.protocolRdd.Pack(data)
	}
	return conn.AsyncWrite(packed)
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
	var (
		err           error
		parsedDataNs  *protocol.ParsedDataNs
		parsedDataRdd *protocol.ParsedDataRdd
		mn            string
	)
	switch s.protocol {
	case ProtocolNs:
		parsedDataNs, err = s.protocolNs.Parse(string(frame))
		if err != nil {
			return
		}
		mn = parsedDataNs.Mn
	case ProtocolRdd:
		parsedDataRdd, err = s.protocolRdd.Parse(string(frame))
		if err != nil {
			return
		}
		mn = parsedDataRdd.Mn
	}
	if mn == "" {
		return
	}
	msg := &Msg{
		data:          frame,
		parsedDataNs:  parsedDataNs,
		parsedDataRdd: parsedDataRdd,
	}
	if s.GetConn(mn) != c {
		s.setConn(mn, c)
		s.handler.OnMn(mn, true)
	}
	s.handler.React(c, msg)
	return
}
