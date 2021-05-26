package fmp

import (
	"sync"
)

type Msg struct {
	mu            sync.RWMutex
	data          []byte
	parsedDataNS  *ParsedDataNS
	parsedDataRdd *ParsedDataRdd
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
