package fmp

import (
	"bytes"

	"github.com/panjf2000/gnet/errors"

	"github.com/panjf2000/gnet"
)

type delimiterCodec struct {
	delimiter []byte
}

func (cc *delimiterCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (cc *delimiterCodec) Decode(c gnet.Conn) ([]byte, error) {
	buf := c.Read()
	idx := bytes.Index(buf, cc.delimiter)

	if idx == -1 {
		return nil, errors.ErrDelimiterNotFound
	}
	c.ShiftN(idx + len(cc.delimiter))
	return buf[:idx], nil
}
