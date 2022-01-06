package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"unsafe"
)

const (
	CHANNEL_PACKET_MAX = 1024
)

var nativeEndian binary.ByteOrder

type Tcp struct {
	fd       int
	conn     net.Conn
	rQueue   chan IPacket
	wQueue   chan IPacket
	packet   func(buf []byte) (IPacket, error)
	buf      []byte
	isClosed bool
}

func init() {
	i := uint16(1)
	if *(*byte)(unsafe.Pointer(&i)) == 0 {
		nativeEndian = binary.BigEndian
	} else {
		nativeEndian = binary.LittleEndian
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
	return &Tcp{fd, conn, make(chan IPacket, CHANNEL_PACKET_MAX), make(chan IPacket, CHANNEL_PACKET_MAX), nil, make([]byte, 0, 2097152), false}
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
			if err == io.EOF {
				return nil, err
			}

			t.buf = append(t.buf, tmp...)
			if n != need {
				continue
			}

			hBuf = t.buf[:hLen]
		}

		bodyLen := hBuf[bLenOffset : bLenOffset+bLen]
		bLength := int(BytesToInt32(bodyLen))
		if bLen > 4 {
			bLength = int(BytesToInt64(bodyLen))
		}

		if l >= (bLength + hLen) {
			packet := make([]byte, bLength+hLen)
			copy(packet, t.buf[:bLength+hLen])
			t.buf = t.buf[bLength+hLen:]
			return packet, nil
		}

		buf := make([]byte, bLength)
		n, err := t.conn.Read(buf)
		if err == io.EOF {
			return nil, err
		}

		if n == 0 {
			continue
		}

		t.buf = append(t.buf, buf...)
	}
}

func (t *Tcp) WQueue() chan IPacket {
	return t.wQueue
}

func (t *Tcp) RQueue() chan IPacket {
	return t.rQueue
}

func (t *Tcp) FD() int {
	return t.fd
}

func (t *Tcp) Write(pack IPacket) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	return t.conn.Write(pack.Serialize())
}

func (t *Tcp) Closed() bool {
	return t.isClosed
}
