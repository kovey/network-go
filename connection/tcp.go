package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/kovey/logger-go/logger"
)

const (
	CHANNEL_PACKET_MAX = 1024
	Packet_Max_Len     = 2097152
)

var nativeEndian binary.ByteOrder = binary.BigEndian

type Tcp struct {
	fd       int
	conn     net.Conn
	rQueue   chan IPacket
	wQueue   chan []byte
	packet   func(buf []byte) (IPacket, error)
	buf      []byte
	isClosed bool
}

func Init(endian string) {
	switch endian {
	case "BigEndian":
		nativeEndian = binary.BigEndian
	case "LittleEndian":
		nativeEndian = binary.LittleEndian
	default:
		nativeEndian = binary.BigEndian
	}
}

func Int32ToBytes(num int32) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, nativeEndian, num)
	return buf.Bytes()
}

func Int64ToBytes(num int64) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, nativeEndian, num)
	return buf.Bytes()
}

func BytesToInt32(buf []byte) int32 {
	b := bytes.NewBuffer(buf)
	var num int32
	binary.Read(b, nativeEndian, &num)
	return num
}

func BytesToInt64(buf []byte) int64 {
	b := bytes.NewBuffer(buf)
	var num int64
	binary.Read(b, nativeEndian, &num)
	return num
}

func NewTcp(fd int, conn net.Conn) *Tcp {
	return &Tcp{fd, conn, make(chan IPacket, CHANNEL_PACKET_MAX), make(chan []byte, CHANNEL_PACKET_MAX), nil, make([]byte, 0, Packet_Max_Len), false}
}

func (t *Tcp) Close() error {
	if t.isClosed {
		return nil
	}

	close(t.rQueue)
	close(t.wQueue)
	t.isClosed = true
	return t.conn.Close()
}

func (t *Tcp) Read(hLen, bLen, bLenOffset int) ([]byte, error) {
	for {
		l := len(t.buf)
		hBuf := make([]byte, hLen)
		if l >= hLen {
			hBuf = t.buf[:hLen]
		} else {
			need := hLen - l
			tmp := make([]byte, need)
			n, err := t.conn.Read(tmp)
			if err != nil {
				return nil, err
			}

			logger.Debug("tmp: %+v", tmp)
			t.buf = append(t.buf, tmp...)
			if n != need {
				continue
			}

			hBuf = t.buf[:hLen]
		}

		bodyLen := hBuf[bLenOffset : bLenOffset+bLen]
		bLength := int(BytesToInt32(bodyLen))
		logger.Debug("length: %d", bLength)
		if bLen > 4 {
			bLength = int(BytesToInt64(bodyLen))
		}

		if bLength < 0 || bLength > Packet_Max_Len-hLen {
			return nil, io.EOF
		}

		if l >= (bLength + hLen) {
			packet := make([]byte, bLength+hLen)
			copy(packet, t.buf[:bLength+hLen])
			t.buf = t.buf[bLength+hLen:]
			return packet, nil
		}

		buf := make([]byte, bLength)
		n, err := t.conn.Read(buf)
		logger.Debug("fianl buf: %+v", buf)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		if l+n > Packet_Max_Len {
			return nil, io.EOF
		}

		t.buf = append(t.buf, buf...)
	}
}

func (t *Tcp) WQueue() chan []byte {
	return t.wQueue
}

func (t *Tcp) RQueue() chan IPacket {
	return t.rQueue
}

func (t *Tcp) FD() int {
	return t.fd
}

func (t *Tcp) Write(pack []byte) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	return t.conn.Write(pack)
}

func (t *Tcp) Send(pack IPacket) error {
	buf := pack.Serialize()
	if buf == nil {
		return fmt.Errorf("pack is nil")
	}

	return t.SendBytes(buf)
}

func (t *Tcp) SendBytes(buf []byte) error {
	if t.isClosed {
		return fmt.Errorf("connection[%d] is closed", t.fd)
	}

	t.wQueue <- buf
	return nil
}

func (t *Tcp) Closed() bool {
	return t.isClosed
}
