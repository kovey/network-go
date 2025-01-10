package connection

type Packet struct {
	Header []byte
	Body   []byte
}

func (p *Packet) Bytes() []byte {
	return append(p.Header, p.Body...)
}
