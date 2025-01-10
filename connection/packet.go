package connection

type Packet struct {
	Header []byte
	Body   []byte
}
