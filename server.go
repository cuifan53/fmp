package fmp

import (
	"errors"
	"sync"

	"github.com/panjf2000/gnet"
)

type Protocol int

const (
	ProtocolNS  Protocol = iota // 国标协议(2017 & 2005)
	ProtocolRdd                 // 远程设备调试协议
)

type EventHandler interface {
	OnOpened(c gnet.Conn)
	OnClosed(c gnet.Conn)
	OnMn(mn string, connect bool)
	React(c gnet.Conn, msg *Msg)
}

func NewServer(port string, protocol Protocol, handler EventHandler) *Server {
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
	protocol Protocol
	handler  EventHandler
}

func (s *Server) Serve() {
	if err := gnet.Serve(
		s,
		s.port,
		gnet.WithMulticore(true),
		gnet.WithCodec(&delimiterCodec{delimiter: s.getDelimiter()}),
	); err != nil {
		panic(err)
	}
}

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

func (s *Server) getDelimiter() []byte {
	var delimiter []byte
	switch s.protocol {
	case ProtocolNS:
		delimiter = []byte(MsgEofNS)
	case ProtocolRdd:
		delimiter = []byte(MsgEofRdd)
	default:
		panic(errors.New("protocol incorrect"))
	}
	return delimiter
}

func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	var err error
	var parsedDataNS *ParsedDataNS
	var parsedDataRdd *ParsedDataRdd
	var mn string

	switch s.protocol {
	case ProtocolNS:
		parsedDataNS, err = parseNS(string(frame))
		if err != nil {
			return
		}
		mn = parsedDataNS.Mn
	case ProtocolRdd:
		parsedDataRdd, err = parseRdd(string(frame))
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
		parsedDataNS:  parsedDataNS,
		parsedDataRdd: parsedDataRdd,
	}
	if s.GetConn(mn) != c {
		s.setConn(mn, c)
		s.handler.OnMn(mn, true)
	}
	s.handler.React(c, msg)
	return
}

func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	s.handler.OnOpened(c)
	return
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for mn, conn := range s.connMap {
		if conn == c {
			delete(s.connMap, mn)
			s.handler.OnMn(mn, false)
		}
	}
	s.handler.OnClosed(c)
	return
}
