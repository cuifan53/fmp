package protocol

import (
	"strconv"
	"strings"
)

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