package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/kovey/debug-go/debug"
)

const (
	CHANNEL_PACKET_MAX = 1024
	Packet_Max_Len     = 2097152
)

var nativeEndian binary.ByteOrder = binary.BigEndian

type Tcp struct {
	fd       uint64
	conn     net.Conn
	wQueue   chan []byte
	buf      []byte
	isClosed bool
	lastTime int64
	ext      map[int64]any
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
	if err := binary.Write(buf, nativeEndian, num); err != nil {
		debug.Erro("int32 to bytes failure, error: %s", err)
	}
	return buf.Bytes()
}

func Int64ToBytes(num int64) []byte {
	buf := bytes.NewBuffer([]byte{})
	if err := binary.Write(buf, nativeEndian, num); err != nil {
		debug.Erro("int64 to bytes failure, error: %s", err)
	}
	return buf.Bytes()
}

func BytesToInt32(buf []byte) int32 {
	b := bytes.NewBuffer(buf)
	var num int32
	if err := binary.Read(b, nativeEndian, &num); err != nil {
		debug.Erro("read int32 from bytes failure, error: %s", err)
	}
	return num
}

func BytesToInt64(buf []byte) int64 {
	b := bytes.NewBuffer(buf)
	var num int64
	if err := binary.Read(b, nativeEndian, &num); err != nil {
		debug.Erro("read int64 from bytes failure, error: %s", err)
	}
	return num
}

func NewTcp(fd uint64, conn net.Conn) *Tcp {
	return &Tcp{
		fd: fd, conn: conn, wQueue: make(chan []byte, CHANNEL_PACKET_MAX), ext: make(map[int64]any),
		buf: make([]byte, 0, Packet_Max_Len), isClosed: false, lastTime: time.Now().Unix(),
	}
}

func (t *Tcp) Close() error {
	if t.isClosed {
		return nil
	}

	t.isClosed = true
	return t.conn.Close()
}

func (t *Tcp) Read(hLen, bLen, bLenOffset int) ([]byte, error) {
	for {
		l := len(t.buf)
		var hBuf []byte
		if l >= hLen {
			hBuf = t.buf[:hLen]
		} else {
			need := hLen - l
			tmp := make([]byte, need)
			n, err := t.conn.Read(tmp)
			if err != nil {
				return nil, err
			}

			if n == 0 {
				continue
			}

			t.buf = append(t.buf, tmp[:n]...)
			if n != need {
				continue
			}

			hBuf = t.buf[:hLen]
			l += n
		}

		bodyLen := hBuf[bLenOffset : bLenOffset+bLen]
		bLength := int(BytesToInt32(bodyLen))
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
			t.lastTime = time.Now().Unix()
			return packet, nil
		}

		buf := make([]byte, bLength)
		n, err := t.conn.Read(buf)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			continue
		}

		if l+n > Packet_Max_Len {
			return nil, io.EOF
		}

		t.buf = append(t.buf, buf[:n]...)
	}
}

func (t *Tcp) WQueue() chan []byte {
	return t.wQueue
}

func (t *Tcp) FD() uint64 {
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

func (t *Tcp) RemoteIp() string {
	addr := t.conn.RemoteAddr().String()
	return strings.Split(addr, ":")[0]
}

func (t *Tcp) Expired() bool {
	return time.Now().Unix() > t.lastTime+60
}

func (t *Tcp) Set(userId int64, val any) {
	t.ext[userId] = val
}

func (t *Tcp) Get(userId int64) (any, bool) {
	val, ok := t.ext[userId]
	return val, ok
}
