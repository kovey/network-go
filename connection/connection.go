package connection

type IConnection interface {
	Read(int, int, int) ([]byte, error)
	Write([]byte) (int, error)
	Send(IPacket) error
	SendBytes([]byte) error
	Close() error
	FD() uint64
	WQueue() <-chan []byte
	Closed() bool
	RemoteIp() string
	Expired() bool
	Set(key int64, value any)
	Get(key int64) (any, bool)
	SQueue() <-chan bool
}

func Get[T any](conn IConnection, key int64) T {
	var res T
	val, ok := conn.Get(key)
	if !ok {
		return res
	}

	tmp, ok := val.(T)
	if !ok {
		return res
	}

	return tmp
}
