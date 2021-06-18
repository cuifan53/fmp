package fmp

import (
	"sync"

	"github.com/cuifan53/fmp/protocol"
)

type Msg struct {
	mu            sync.RWMutex
	data          []byte
	parsedDataNs  *protocol.ParsedDataNs
	parsedDataRdd *protocol.ParsedDataRdd
}

func (m *Msg) GetData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

func (m *Msg) GetParsedDataNs() *protocol.ParsedDataNs {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataNs
}

func (m *Msg) GetParsedDataRdd() *protocol.ParsedDataRdd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedDataRdd
}
