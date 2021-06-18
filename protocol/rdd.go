package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Rdd struct {
	name       string
	header     string
	dataLenLen int
	crcLen     int
	eof        string
}

// Parse 解析tcp数据包
func (p *Rdd) Parse(originMsg string) (interface{}, error) {
	parsedData := ParsedDataRdd{
		OriginMsg: originMsg,
	}
	dataAndCrc := originMsg[len(p.header)+p.dataLenLen:]
	data := dataAndCrc[:len(dataAndCrc)-p.crcLen]
	msgCrc := dataAndCrc[len(data):]
	// crc校验
	realCrc := crc(data)
	if msgCrc != realCrc {
		return nil, errors.New("crc校验失败")
	}
	// 按CP数据段分割
	cpIndex := strings.Index(data, "CP=&&")
	// 编码区
	code := data[:cpIndex] // MN=WXTC20191121196
	p.parseCode(&parsedData, code)
	// CP区
	cp := data[cpIndex+5:]
	cp = cp[:len(cp)-2] // 这里的2是字符串最后的2个&&
	if cp != "" {
		if err := p.parseCp(&parsedData, cp); err != nil {
			return nil, err
		}
	}
	return &parsedData, nil
}

// 解析编码区
func (p *Rdd) parseCode(parsedData *ParsedDataRdd, code string) {
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
func (p *Rdd) parseCp(parsedData *ParsedDataRdd, cp string) error {
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

func (p *Rdd) Pack(data string) []byte {
	dataLenStr := strconv.Itoa(len(data))
	header := p.header + ("00000000" + dataLenStr)[len(dataLenStr):]
	crcData := crc(data)
	return []byte(header + data + crcData + p.eof)
}

func (p *Rdd) Eof() []byte {
	return []byte(p.eof)
}

func (p *Rdd) Name() string {
	return p.name
}
