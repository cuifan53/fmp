package fmp

import (
	"errors"
	"sync"
	"time"

	"github.com/cuifan53/fmp/protocol"
	"github.com/panjf2000/gnet"
)

func NewServer(port string, timeout time.Duration, protocolName protocol.ProtocolName, handler IEventHandler) *Server {
	return &Server{
		timeout:  timeout,
		port:     port,
		protocol: protocol.NewProtocol(protocolName),
		handler:  handler,
	}
}

type Server struct {
	*gnet.EventServer
	connMap       sync.Map // map[string]gnet.Conn
	connLatestMap sync.Map // map[gnet.Conn]time.Time
	timeout       time.Duration
	port          string
	protocol      protocol.IProtocol
	handler       IEventHandler
}

func (s *Server) Serve() {
	go func() {
		for {
			s.connLatestMap.Range(func(c, v interface{}) bool {
				if time.Now().After(v.(time.Time).Add(s.timeout)) {
					_ = c.(gnet.Conn).Close()
					s.connLatestMap.Delete(c)
				}
				return true
			})
			time.Sleep(time.Second)
		}
	}()
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
	mns := make([]string, 0)
	s.connMap.Range(func(mn, _ interface{}) bool {
		if mn != "" {
			mns = append(mns, mn.(string))
		}
		return true
	})
	return mns
}

// GetConn 通过MN获取Tcp连接
func (s *Server) GetConn(mn string) gnet.Conn {
	conn, ok := s.connMap.Load(mn)
	if !ok {
		return nil
	}
	return conn.(gnet.Conn)
}

func (s *Server) setConn(mn string, conn gnet.Conn) {
	s.connMap.Store(mn, conn)
}

func (s *Server) removeConn(mn string) {
	s.connMap.Delete(mn)
}

// Reset 重置服务器 断开所有链接
func (s *Server) Reset() {
	s.connMap.Range(func(mn, conn interface{}) bool {
		_ = conn.(gnet.Conn).Close()
		s.connMap.Delete(mn)
		return true
	})
}

// ** 以下为重写gnet.EventServer方法 ** //

func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	s.connLatestMap.Store(c, time.Now())
	s.handler.OnOpened(c)
	return
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	s.connMap.Range(func(mn, conn interface{}) bool {
		if conn == c {
			s.connMap.Delete(mn)
			s.handler.OnMn(mn.(string), false)
		}
		return true
	})
	s.handler.OnClosed(c)
	return
}

func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	s.connLatestMap.Store(c, time.Now())
	parsedData, err := s.protocol.Parse(frame)
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
	case protocol.ProtocolNameTc:
		parsedDataTc, ok := parsedData.(*protocol.ParsedDataTc)
		if !ok {
			return
		}
		msg.parsedDataTc = parsedDataTc
		mn = parsedDataTc.Header.Token
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
