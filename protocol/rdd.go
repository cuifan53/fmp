package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/cuifan53/fmp"
)

const (
	MsgHeaderRdd     = "##**"
	MsgDataLenLenRdd = 8
	MsgCrcLenRdd     = 4
	MsgEofRdd        = "**\r\n"
)

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

func PackRdd(data string) []byte {
	dataLenStr := strconv.Itoa(len(data))
	header := MsgHeaderRdd + ("00000000" + dataLenStr)[len(dataLenStr):]
	crcData := fmp.Crc(data)
	return []byte(header + data + crcData + MsgEofRdd)
}

// ParseRdd 解析tcp数据包
func ParseRdd(originMsg string) (*ParsedDataRdd, error) {
	parsedData := ParsedDataRdd{
		OriginMsg: originMsg,
	}
	dataAndCrc := originMsg[len(MsgHeaderRdd)+MsgDataLenLenRdd:]
	data := dataAndCrc[:len(dataAndCrc)-MsgCrcLenRdd]
	msgCrc := dataAndCrc[len(data):]
	// crc校验
	realCrc := fmp.Crc(data)
	if msgCrc != realCrc {
		return nil, errors.New("crc校验失败")
	}
	// 按CP数据段分割
	cpIndex := strings.Index(data, "CP=&&")
	// 编码区
	code := data[:cpIndex] // MN=WXTC20191121196
	parseCodeRdd(&parsedData, code)
	// CP区
	cp := data[cpIndex+5:]
	cp = cp[:len(cp)-2] // 这里的2是字符串最后的2个&&
	if cp != "" {
		if err := parseCpRdd(&parsedData, cp); err != nil {
			return nil, err
		}
	}
	return &parsedData, nil
}

// 解析编码区
func parseCodeRdd(parsedData *ParsedDataRdd, code string) {
	m := make(map[string]string)
	tmp := strings.Split(code, ";") // ["MN=WXTC20191121196"]
	for _, v := range tmp {
		if !strings.Contains(v, "=") {
			continue
		}
		tmp1 := strings.Split(v, "=") // ["MN", WXTC20191121196]
		m[tmp1[0]] = tmp1[1]
	}
	parsedData.Mn = m["MN"]
}

// 解析CP区
func parseCpRdd(parsedData *ParsedDataRdd, cp string) error {
	parsedData.Cp = cp
	type CpBody struct {
		Cmd      string `json:"cmd"`
		CmdId    string `json:"cmdId"`
		CmdStata string `json:"cmdStata"`
		RepParam string `json:"repParam"`
	}
	cpBody := CpBody{}
	if err := json.Unmarshal([]byte(cp), &cpBody); err != nil {
		return err
	}
	parsedData.Cmd = cpBody.Cmd
	parsedData.CmdId = cpBody.CmdId
	parsedData.CmdStata = cpBody.CmdStata
	if cpBody.RepParam != "" {
		repParam := RepParamRdd{}
		if err := json.Unmarshal([]byte(cpBody.RepParam), &repParam); err != nil {
			return err
		}
		parsedData.RepParam = &repParam
	}
	return nil
}
