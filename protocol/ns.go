package protocol

import (
	"errors"
	"strconv"
	"strings"
)

type Ns struct {
	name       string
	header     string
	dataLenLen int
	crcLen     int
	eof        string
}

// Parse 解析tcp数据包
func (p *Ns) Parse(originMsg string) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	// 检测长度
	if len(originMsg) < len(p.header)+p.dataLenLen+p.crcLen+len(p.eof) {
		return nil, errors.New("报文长度不足")
	}

	// 检测包头
	if originMsg[0:len(p.header)] != p.header {
		return nil, errors.New("报文包头异常")
	}

	parsedData := ParsedDataNs{
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
	tmp := strings.Split(data, "CP=&&")
	// 编码区
	code := tmp[0] // ST=32;CN=2011;PW=123456;MN=WXTC20191121196;Flag=0;
	p.parseCode(&parsedData, code)
	// CP区
	cp := tmp[1][:len(tmp[1])-2] // 这里的2是字符串最后的2个&&
	p.parseCp(&parsedData, cp)
	// 解析Flag
	p.parseFlag(&parsedData)
	return &parsedData, nil
}

// 解析编码区
func (p *Ns) parseCode(parsedData *ParsedDataNs, code string) {
	m := make(map[string]string)
	tmp := strings.Split(code, ";") // ["ST=32", "CN=2011"]
	for _, v := range tmp {
		if !strings.Contains(v, "=") {
			continue
		}
		tmp1 := strings.Split(v, "=") // ["ST", 32]
		m[tmp1[0]] = tmp1[1]
	}
	parsedData.Qn = m["QN"]
	parsedData.St = m["ST"]
	parsedData.Cn = m["CN"]
	parsedData.Pw = m["PW"]
	parsedData.Mn = m["MN"]
	flag, _ := strconv.Atoi(m["Flag"])
	parsedData.Flag = flag
	pnum, _ := strconv.Atoi(m["PNUM"])
	parsedData.Pnum = pnum
	pno, _ := strconv.Atoi(m["PNO"])
	parsedData.Pno = pno
}

// 解析CP数据区
func (p *Ns) parseCp(parsedData *ParsedDataNs, cp string) {
	m := make(map[string]string)
	tmp := strings.Split(cp, ";") // ["DataTime=20200114120000", "011-Rtd=0,011-Flag=B"]
	for _, v := range tmp {
		if !strings.Contains(v, "=") {
			continue
		}
		if !strings.Contains(v, ",") { // DataTime=20200114120000
			tmp1 := strings.Split(v, "=") // ["DataTime", 20200114120000]
			m[tmp1[0]] = tmp1[1]
		} else { // 011-Rtd=0,011-Flag=B
			tmp2 := strings.Split(v, ",") // ["011-Rtd=0", "011-Flag=B"]
			for _, v1 := range tmp2 {
				if !strings.Contains(v1, "=") {
					continue
				}
				tmp3 := strings.Split(v1, "=") // ["011-Rtd", 0]
				m[tmp3[0]] = tmp3[1]
			}
		}
	}
	parsedData.Cp = m
}

// 解析Flag
// 规则 8位  000000代表2005版 000001代表2017版 第7位为是否有数据包序号 即是否有 PNUM PNO包号 第8位为是否需要应答
// 如：00000111 十进制7 代表2017版 有数据包序号 需要应答
// 故当前所有Flag如下
// 十进制  二进制      含义
// 0      00000000   2005版 无数据包序号 无需应答
// 1      00000001   2005版 无数据包序号 需应答
// 2      00000010   2005版 有数据包序号 无需应答
// 3      00000011   2005版 有数据包序号 需应答
// 4      00000100   2017版 无数据包序号 无需应答
// 5      00000101   2017版 无数据包序号 需应答
// 6      00000110   2017版 有数据包序号 无需应答
// 7      00000111   2017版 有数据包序号 需应答
// 8      00001000   2017扩展版 无数据包序号 无需应答
// 9      00001001   2017扩展版 无数据包序号 需应答
// 10     00001010   2017扩展版 有数据包序号 无需应答
// 11     00001011   2017扩展版 有数据包序号 需应答
func (p *Ns) parseFlag(parsedData *ParsedDataNs) {
	bFlag := strconv.FormatInt(int64(parsedData.Flag), 2)
	bFlag = ("00000000" + bFlag)[len(bFlag):]
	if bFlag[4:5] == "1" || bFlag[5:6] == "1" {
		parsedData.Protocol = "2017"
	} else {
		parsedData.Protocol = "2005"
	}
	if bFlag[len(bFlag)-1:] == "1" {
		parsedData.NeedReply = true
	} else {
		parsedData.NeedReply = false
	}
}

func (p *Ns) Pack(data string) []byte {
	dataLenStr := strconv.Itoa(len(data))
	header := p.header + ("0000" + dataLenStr)[len(dataLenStr):]
	crcData := crc(data)
	return []byte(header + data + crcData + p.eof)
}

func (p *Ns) Eof() []byte {
	return []byte(p.eof)
}

func (p *Ns) Name() string {
	return p.name
}
