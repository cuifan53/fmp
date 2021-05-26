package fmp

import (
	"sync"

	"github.com/panjf2000/gnet"
)

type Msg struct {
	mu            sync.RWMutex
	conn          gnet.Conn
	data          []byte
	parsedDataNS  *ParsedDataNS
	parsedDataRdd *ParsedDataRdd
}

func (m *Msg) GetConn() gnet.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *Msg) GetData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

func (m *Msg) GetParsedDataNS() *ParsedDataNS {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataNS
}

func (m *Msg) GetParsedDataRdd() *ParsedDataRdd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataRdd
}
