package connection

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	headerLen     int
	bodyLenOffset int
	bodyLengthLen int
	bodyLenType   LenType
	endian        binary.ByteOrder
}

func NewHeader() *Header {
	return &Header{bodyLenType: Len_Type_Int32, bodyLengthLen: 4, endian: binary.BigEndian, headerLen: 4, bodyLenOffset: 0}
}

func (h *Header) WithBodyLenType(t LenType) *Header {
	h.bodyLenType = t
	h.bodyLengthLen = h._len(t)
	return h
}

func (h *Header) WithBodyLenOffset(bodyLenOffset int) *Header {
	h.bodyLenOffset = bodyLenOffset
	return h
}

func (h *Header) WithHeaderLen(headerLen int) *Header {
	h.headerLen = headerLen
	return h
}

func (h *Header) WithEndian(endian binary.ByteOrder) *Header {
	h.endian = endian
	return h
}

func (h *Header) BodyLenOffset() int {
	return h.bodyLenOffset
}

func (h *Header) _len(t LenType) int {
	switch t {
	case Len_Type_Int8, Len_Type_UInt8:
		return 1
	case Len_Type_Int16, Len_Type_UInt16:
		return 2
	case Len_Type_Int32, Len_Type_UInt32:
		return 4
	case Len_Type_Int64, Len_Type_UInt64:
		return 8
	default:
		return 4
	}
}

func (h *Header) Header(bodyLen int) []byte {
	var buffer bytes.Buffer
	headers := make([]byte, h.headerLen)
	switch h.bodyLenType {
	case Len_Type_Int8:
		binary.Write(&buffer, h.endian, int8(bodyLen))
		headers[h.bodyLenOffset] = buffer.Bytes()[0]
	case Len_Type_Int16:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, int16(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
	case Len_Type_Int32:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, int32(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
		headers[h.bodyLenOffset+2] = bs[2]
		headers[h.bodyLenOffset+3] = bs[3]
	case Len_Type_Int64:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, int64(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
		headers[h.bodyLenOffset+2] = bs[2]
		headers[h.bodyLenOffset+3] = bs[3]
		headers[h.bodyLenOffset+4] = bs[4]
		headers[h.bodyLenOffset+5] = bs[5]
		headers[h.bodyLenOffset+6] = bs[6]
		headers[h.bodyLenOffset+7] = bs[7]
	case Len_Type_UInt8:
		headers[h.bodyLenOffset] = byte(bodyLen)
	case Len_Type_UInt16:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, uint16(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
	case Len_Type_UInt32:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, uint32(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
		headers[h.bodyLenOffset+2] = bs[2]
		headers[h.bodyLenOffset+3] = bs[3]
	case Len_Type_UInt64:
		var buffer bytes.Buffer
		binary.Write(&buffer, h.endian, uint64(bodyLen))
		bs := buffer.Bytes()
		headers[h.bodyLenOffset] = bs[0]
		headers[h.bodyLenOffset+1] = bs[1]
		headers[h.bodyLenOffset+2] = bs[2]
		headers[h.bodyLenOffset+3] = bs[3]
		headers[h.bodyLenOffset+4] = bs[4]
		headers[h.bodyLenOffset+5] = bs[5]
		headers[h.bodyLenOffset+6] = bs[6]
		headers[h.bodyLenOffset+7] = bs[7]
	}
	return headers
}
