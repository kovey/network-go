package connection

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"
)

var Err_Closed = errors.New("connection is closed")
var Err_Unkown_Body_Len_Type = errors.New("unkown body length type")
var Err_Packet_Out_Range = errors.New("packet out of range")

type LenType byte

const (
	Len_Type_Int8   LenType = 1
	Len_Type_Int16  LenType = 2
	Len_Type_Int32  LenType = 3
	Len_Type_Int64  LenType = 4
	Len_Type_UInt8  LenType = 5
	Len_Type_UInt16 LenType = 6
	Len_Type_UInt32 LenType = 7
	Len_Type_UInt64 LenType = 8
)

type Connection struct {
	conn           net.Conn
	header         *Header
	maxLen         int
	packBuff       []byte
	readLen        int
	fd             uint64
	isClosed       bool
	connectTime    int64         // nano seconds
	lastActiveTime int64         // nano seconds
	maxIdleTime    time.Duration // max idle time
	packets        chan *Packet
}

func NewConnection(fd uint64, conn net.Conn) *Connection {
	return NewConnectionBy(NewHeader(), fd, conn)
}

func NewConnectionBy(header *Header, fd uint64, conn net.Conn) *Connection {
	now := time.Now()
	return &Connection{conn: conn, maxLen: 8192, header: header, fd: fd, connectTime: now.UnixNano(), lastActiveTime: now.UnixNano(), packets: make(chan *Packet, 1024)}
}

func (c *Connection) Header() *Header {
	return c.header
}

func (c *Connection) WithConn(conn net.Conn) *Connection {
	c.readLen = 0
	c.conn = conn
	return c
}

func (c *Connection) FD() uint64 {
	return c.fd
}

func (c *Connection) Expired(now time.Time) bool {
	if c.maxIdleTime <= 0 {
		return false
	}

	return now.UnixNano() > c.lastActiveTime+int64(c.maxIdleTime)
}

func (c *Connection) WithMaxIdleTime(maxIdleTime time.Duration) *Connection {
	c.maxIdleTime = maxIdleTime
	return c
}

func (c *Connection) WithBodyLenType(t LenType) *Connection {
	c.header.WithBodyLenType(t)
	return c
}

func (c *Connection) WithEndian(e binary.ByteOrder) *Connection {
	c.header.WithEndian(e)
	return c
}

func (c *Connection) WithMaxLen(maxLen int) *Connection {
	c.maxLen = maxLen
	c.packBuff = make([]byte, c.maxLen)
	return c
}

func (c *Connection) WithHeaderLen(length int) *Connection {
	c.header.WithHeaderLen(length)
	return c
}

func (c *Connection) WithBodyLenOffset(offset int) *Connection {
	c.header.WithBodyLenOffset(offset)
	return c
}

func (c *Connection) Write(data []byte) error {
	if c.isClosed {
		return Err_Closed
	}

	_, err := c.conn.Write(data)
	return err
}

func (c *Connection) ReadLoop() {
	for {
		packet, err := c.Read()
		if err != nil {
			break
		}

		c.packets <- packet
	}
}

func (c *Connection) Packets() <-chan *Packet {
	return c.packets
}

func (c *Connection) Read() (*Packet, error) {
	if c.packBuff == nil {
		c.packBuff = make([]byte, c.maxLen)
	}
	var bodyLen = 0
	var err error
	for {
		if bodyLen == 0 {
			if c.readLen >= c.header.headerLen {
				bodyLen, err = c.bodyLen(c.packBuff[c.header.bodyLenOffset : c.header.bodyLenOffset+c.header.bodyLengthLen])
				if err != nil {
					return nil, err
				}
			}
		}

		if c.readLen >= c.header.headerLen+bodyLen {
			return c.copyBuff(bodyLen), nil
		}

		if c.readLen >= c.maxLen {
			return nil, Err_Packet_Out_Range
		}

		n, err := c.conn.Read(c.packBuff[c.readLen:])
		if err != nil {
			return nil, err
		}

		c.readLen += n
	}
}

func (c *Connection) copyBuff(bodyLen int) *Packet {
	buffLen := c.header.headerLen + bodyLen
	p := &Packet{Body: make([]byte, bodyLen), Header: make([]byte, c.header.headerLen)}
	copy(p.Header, c.packBuff[:c.header.headerLen])
	copy(p.Body, c.packBuff[c.header.headerLen:])
	copy(c.packBuff, c.packBuff[buffLen:c.readLen])
	c.readLen -= buffLen

	c.lastActiveTime = time.Now().UnixNano()
	return p
}

func (c *Connection) bodyLen(data []byte) (int, error) {
	buffer := bytes.NewBuffer(data)
	switch c.header.bodyLenType {
	case Len_Type_Int8:
		var l int8
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_Int16:
		var l int16
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_Int32:
		var l int32
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_Int64:
		var l int64
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_UInt8:
		var l uint8
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_UInt16:
		var l uint16
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_UInt32:
		var l uint32
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	case Len_Type_UInt64:
		var l uint64
		err := binary.Read(buffer, c.header.endian, &l)
		return int(l), err
	}

	return 0, Err_Unkown_Body_Len_Type
}

func (c *Connection) Close() error {
	if c.isClosed {
		return Err_Closed
	}

	c.isClosed = true
	return c.conn.Close()
}
