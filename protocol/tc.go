package protocol

import (
	"encoding/json"
)

type Tc struct {
	name string
	eof  string
}

// Parse 解析tcp数据包
func (p *Tc) Parse(frame []byte) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	var parsedData ParsedDataTc
	if err := json.Unmarshal(frame, &parsedData); err != nil {
		return nil, err
	}
	return &parsedData, nil
}

func (p *Tc) Pack(data string) []byte {
	return []byte(data + p.eof)
}

func (p *Tc) Eof() []byte {
	return []byte(p.eof)
}

func (p *Tc) Name() string {
	return p.name
}
