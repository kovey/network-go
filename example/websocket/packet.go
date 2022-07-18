package websocket

import (
	"encoding/json"
)

type Packet struct {
	Action int    `json:"action"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
}

func (p *Packet) Serialize() []byte {
	info, err := json.Marshal(p)
	if err != nil {
		return nil
	}

	return info
}

func (p *Packet) Unserialize(buf []byte) error {
	return json.Unmarshal(buf, p)
}
