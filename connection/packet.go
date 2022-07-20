package connection

type IPacket interface {
	Serialize() []byte
	Unserialize([]byte) error
}

type PacketConfig struct {
	HeaderLength  int    `json:"header_length" yaml:"header_length"`
	BodyLenOffset int    `json:"body_length_offset" yaml:"body_length_offset"`
	BodyLenLen    int    `json:"body_length_len" yaml:"body_length_len"`
	Endian        string `json:"endian" yaml:"endian"`
}

type Default struct {
	lenOffset  int
	bodyOffset int
	buf        []byte
}

func NewDefault() *Default {
	return &Default{lenOffset: 0, bodyOffset: 4}
}

func (d *Default) Serialize() []byte {
	return nil
}

func (d *Default) Unserialize(buf []byte) error {
	d.buf = buf
	return nil
}

func (d *Default) LenOffset() int {
	return d.lenOffset
}

func (d *Default) BodyOffset() int {
	return d.bodyOffset
}
