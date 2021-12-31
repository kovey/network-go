package connection

type IConnection interface {
	Read(int, int, int) ([]byte, error)
	Write(IPacket) (int, error)
	Close() error
	FD() int
	RQueue() chan IPacket
	WQueue() chan IPacket
	Closed() bool
}
