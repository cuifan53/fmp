package fmp

import (
	"sync"

	"github.com/cuifan53/fmp/protocol"
)

type Msg struct {
	mu            sync.RWMutex
	data          []byte
	parsedDataNS  *protocol.ParsedDataNS
	parsedDataRdd *protocol.ParsedDataRdd
}

func (m *Msg) GetData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

func (m *Msg) GetParsedDataNS() *protocol.ParsedDataNS {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataNS
}

func (m *Msg) GetParsedDataRdd() *protocol.ParsedDataRdd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataRdd
}
