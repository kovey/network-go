package connection

type Packet struct {
	Header []byte
	Body   []byte
}

func (p *Packet) Bytes() []byte {
	return append(p.Header, p.Body...)
}

func NewPacket(body []byte, header *Header) *Packet {
	p := &Packet{Body: body}
	p.Header = header.Header(len(body))
	return p
}
