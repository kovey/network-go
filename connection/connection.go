package connection

type IConnection interface {
	Read(int, int, int) ([]byte, error)
	Write([]byte) (int, error)
	Send(IPacket) error
	SendBytes([]byte) error
	Close() error
	FD() int
	RQueue() chan IPacket
	WQueue() chan []byte
	Closed() bool
}
