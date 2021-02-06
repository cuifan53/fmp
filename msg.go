package fmp

import "sync"

const (
	MsgHeader     = "##"
	MsgEof        = "\r\n"
	MsgHeaderLen  = 2 // 包头2
	MsgDataLenLen = 4 // 数据段长度4
	MsgCrcLen     = 4 // crc长度4
	MsgEofLen     = 2 // eof \r\n 长度2
)

type Msg struct {
	conn    *Conn
	data    []byte
	dataMap map[string]string
	mu      sync.RWMutex
}

func (m *Msg) GetConn() *Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *Msg) GetDataMap() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dataMap
}
