package fmp

import (
	"strconv"
	"strings"
)

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
