package fmp

const (
	MsgHeader     = "##"
	MsgEof        = "\r\n"
	MsgHeaderLen  = 2 // 包头2
	MsgDataLenLen = 4 // 数据段长度4
	MsgCrcLen     = 4 // crc长度4
	MsgEofLen     = 2 // eof \r\n 长度2
)

type Msg struct {
	Data    []byte
	DataMap map[string]string
}