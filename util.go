package fmp

import (
	"errors"
	"strconv"
	"strings"
)

var U = &util{}

type util struct{}

func (u *util) Pack(data string) string {
	dataLenStr := strconv.Itoa(len(data))
	header := MsgHeader + ("0000" + dataLenStr)[len(dataLenStr):]
	crc := u.crc(data)
	return header + data + crc + MsgEof
}

func (*util) crc(s string) string {
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

// 解析tcp数据包
func (u *util) parse(originMsg string) (map[string]string, error) {
	ret := make(map[string]string)
	// 原始报文
	ret["OriginMsg"] = originMsg
	dataAndCrc := originMsg[MsgHeaderLen+MsgDataLenLen:]
	data := dataAndCrc[:len(dataAndCrc)-MsgCrcLen]
	crc := dataAndCrc[len(data):]
	// crc校验
	realCrc := u.crc(data)
	if crc != realCrc {
		return nil, errors.New("crc校验失败")
	}
	// 按CP数据段分割
	tmp := strings.Split(data, "CP=&&")
	// 编码区
	code := tmp[0] // ST=32;CN=2011;PW=123456;MN=WXTC20191121196;Flag=0;
	u.parseCode(ret, code)
	// CP区
	cp := tmp[1][:len(tmp[1])-2] // 这里的2是字符串最后的2个&&
	ret["CP"] = cp
	u.parseCp(ret, cp)
	// 解析Flag
	u.parseFlag(ret)
	// 数据校验
	if err := u.validate(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// 解析编码区
func (*util) parseCode(ret map[string]string, code string) {
	tmp := strings.Split(code, ";") // ["ST=32", "CN=2011"]
	for _, v := range tmp {
		if !strings.Contains(v, "=") {
			continue
		}
		tmp1 := strings.Split(v, "=") // ["ST", 32]
		ret[tmp1[0]] = tmp1[1]
	}
}

// 解析CP数据区
func (*util) parseCp(ret map[string]string, cp string) {
	tmp := strings.Split(cp, ";") // ["DataTime=20200114120000", "011-Rtd=0,011-Flag=B"]
	for _, v := range tmp {
		if !strings.Contains(v, "=") {
			continue
		}
		if !strings.Contains(v, ",") { // DataTime=20200114120000
			tmp1 := strings.Split(v, "=") // ["DataTime", 20200114120000]
			ret[tmp1[0]] = tmp1[1]
		} else { // 011-Rtd=0,011-Flag=B
			tmp2 := strings.Split(v, ",") // ["011-Rtd=0", "011-Flag=B"]
			for _, v1 := range tmp2 {
				if !strings.Contains(v1, "=") {
					continue
				}
				tmp3 := strings.Split(v1, "=") // ["011-Rtd", 0]
				ret[tmp3[0]] = tmp3[1]
			}
		}
	}
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
func (*util) parseFlag(ret map[string]string) {
	if flag, ok := ret["Flag"]; ok {
		flagInt64, _ := strconv.ParseInt(flag, 10, 64)
		bFlag := strconv.FormatInt(flagInt64, 2)
		bFlag = ("00000000" + bFlag)[len(bFlag):]
		if bFlag[4:5] == "1" || bFlag[5:6] == "1" {
			ret["Protocol"] = "2017"
		} else {
			ret["Protocol"] = "2005"
		}
		if bFlag[len(bFlag)-2:len(bFlag)-1] == "1" {
			ret["HasPN"] = "1"
		} else {
			ret["HasPN"] = "0"
		}
		if bFlag[len(bFlag)-1:] == "1" {
			ret["NeedReply"] = "1"
		} else {
			ret["NeedReply"] = "0"
		}
	} else {
		ret["Protocol"] = "2005"
		ret["HasPN"] = "0"
		ret["NeedReply"] = "0"
	}
}

// 校验数据
func (*util) validate(ret map[string]string) error {
	if _, ok := ret["ST"]; !ok {
		return errors.New("ST字段不存在")
	}
	if _, ok := ret["CN"]; !ok {
		return errors.New("CN字段不存在")
	}
	if _, ok := ret["MN"]; !ok {
		return errors.New("MN字段不存在")
	}
	return nil
}
