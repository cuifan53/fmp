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

type ParsedData struct {
	Qn   string `json:"qn"`   // 请求编码
	St   string `json:"st"`   // 系统编码
	Cn   string `json:"cn"`   // 命令编码
	Pw   string `json:"pw"`   // 访问密码
	Mn   string `json:"mn"`   // 设备唯一标识
	Flag int    `json:"flag"` // 标志位
	Pnum string `json:"pnum"` // 总包数
	Pno  string `json:"pno"`  // 当前数据包包号
	Cp   string `json:"cp"`   // 数据区

	OriginMsg string            `json:"originMsg"` // 原始报文 不包含EOF结尾
	ParsedCp  map[string]string `json:"parsedCp"`  // 解析数据区
	Protocol  string            `json:"protocol"`  // 报文协议 2017 | 2005
	HasPno    bool              `json:"hasPno"`    // 是否有包号(拆包)
	NeedReply bool              `json:"needReply"` // 是否需要应答
}

type Msg struct {
	conn       *Conn
	data       []byte
	parsedData *ParsedData
	mu         sync.RWMutex
}

func (m *Msg) GetConn() *Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conn
}

func (m *Msg) GetData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data
}

func (m *Msg) GetParsedData() *ParsedData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parsedData
}
