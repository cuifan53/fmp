package protocol

import (
	"strconv"
	"strings"
)

type IProtocol interface {
	Parse(frame []byte) (interface{}, error)
	Pack(data string) []byte
	Eof() []byte
	Name() string
}

// ProtocolName 协议名称
type ProtocolName string

const (
	ProtocolNameNs  = "Ns"  // 国标协议(2017 & 2005)
	ProtocolNameRdd = "Rdd" // 远程设备调试协议
	ProtocolNameTc  = "Tc"  // 天创伟业自定义协议
)

// NewProtocol 创建协议
func NewProtocol(protocolName ProtocolName) IProtocol {
	switch protocolName {
	case ProtocolNameNs:
		return &Ns{
			name:       ProtocolNameNs,
			header:     "##",
			dataLenLen: 4,
			crcLen:     4,
			eof:        "\r\n",
		}
	case ProtocolNameRdd:
		return &Rdd{
			name:       ProtocolNameRdd,
			header:     "##**",
			dataLenLen: 8,
			crcLen:     4,
			eof:        "**\r\n",
		}
	case ProtocolNameTc:
		return &Tc{
			name: ProtocolNameTc,
			eof:  "\r\n",
		}
	default:
		return nil
	}
}

// ParsedDataNs Ns协议解析数据
type ParsedDataNs struct {
	OriginMsg string            `json:"originMsg"` // 原始报文 不包含EOF结尾
	Qn        string            `json:"qn"`        // 请求编码
	St        string            `json:"st"`        // 系统编码
	Cn        string            `json:"cn"`        // 命令编码
	Pw        string            `json:"pw"`        // 访问密码
	Mn        string            `json:"mn"`        // 设备唯一标识
	Flag      int               `json:"flag"`      // 标志位
	Pnum      int               `json:"pnum"`      // 总包数
	Pno       int               `json:"pno"`       // 当前数据包包号
	Cp        map[string]string `json:"cp"`        // 数据区
	Protocol  string            `json:"protocol"`  // 报文协议 2017 | 2005
	NeedReply bool              `json:"needReply"` // 是否需要应答
}

// ParsedDataRdd Rdd协议解析数据
type ParsedDataRdd struct {
	OriginMsg string       `json:"originMsg"` // 原始报文 不包含EOF结尾
	Mn        string       `json:"mn"`        // 设备唯一标识
	Cp        string       `json:"cp"`        // 数据区
	Cmd       string       `json:"cmd"`       // 指令
	CmdId     string       `json:"cmdId"`     // 指令id
	CmdStata  string       `json:"cmdStata"`  // 指令状态 Doing End
	RepParam  *RepParamRdd `json:"repParam"`  // 回包内容
}
type RepParamRdd struct {
	RepCode      string `json:"repCode"` // 回包码
	RepStat      string `json:"repStat"` // 回包状态 Success Fail
	RepSendParam string `json:"repSendParam"`
}

// ParsedDataTc Tc协议解析数据
type ParsedDataTc struct {
	Header TcHeader `json:"header"`
	Body   TcBody   `json:"body"`
}
type TcHeader struct {
	Sequence      int    `json:"sequence"`       // 操作序列号
	Timestamp     int    `json:"timestamp"`      // 时间戳
	Token         string `json:"token"`          // 设备编号
	Id            int    `json:"id"`             // 数据所属系统id
	MessageType   string `json:"message.type"`   // 消息类型
	MessageLength int    `json:"message.length"` // 消息长度
}
type TcBody struct {
	Length  int                    `json:"length"`  // 当前内容长度
	Flag    int                    `json:"flag"`    // 是否压缩 0未压缩 1已压缩
	Content map[string]interface{} `json:"content"` // 内容
}

// crc计算
func crc(s string) string {
	ret := 0xFFFF
	num := 0xA001
	inum := 0
	sb := make([]int, 0)
	for i := 0; i < len(s); i++ {
		sb = append(sb, int(rune(s[i])))
	}
	for i := 0; i < len(sb); i++ {
		inum = sb[i]
		ret = (ret >> 8) & 0x00FF
		ret ^= inum
		for j := 0; j < 8; j++ {
			flag := ret % 2
			ret = ret >> 1
			if flag == 1 {
				ret = ret ^ num
			}
		}
	}
	retHexStr := strings.ToUpper(strconv.FormatInt(int64(ret), 16))
	return ("0000" + retHexStr)[len(retHexStr):]
}
