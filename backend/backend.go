package backend

import (
	"math"
	"errors"
	"encoding/binary"
)

var ErrTruncatedData = errors.New("cod: unmarshal encountered truncated data")

const (
	sizeUint8 = 1
	sizeUint16 = 2
	sizeUint32 = 4
	sizeUint64 = 8
	sizeInt8 = 1
	sizeInt16 = 2
	sizeInt32 = 4
	sizeInt64 = 8
)

//TODO: could add another directory level so that you can add different implementations to test against eachother

// TODOs:
// 1. Measure varint encoding vs fixed int encoding

var order = binary.LittleEndian

func WriteUint(bs []byte, v uint) []byte {
	return binary.AppendUvarint(bs, uint64(v))
}
func ReadUint(bs []byte) (uint, int, error) {
	val, n := binary.Uvarint(bs)
	if n <= 0 { return 0, 0, ErrTruncatedData }
	return uint(val), n, nil
}
func WriteInt(bs []byte, v int) []byte {
	return binary.AppendVarint(bs, int64(v))
}
func ReadInt(bs []byte) (int, int, error) {
	val, n := binary.Varint(bs)
	if n <= 0 { return 0, 0, ErrTruncatedData }
	return int(val), n, nil
}

//--------------------------------------------------------------------------------
// Read/Append - Fixed Width
//--------------------------------------------------------------------------------
// Unsigned Integers
func WriteUint8(bs []byte, v uint8) []byte {
	return append(bs, v)
}

func WriteUint16(bs []byte, v uint16) []byte {
	return order.AppendUint16(bs, v)
}

func WriteUint32(bs []byte, v uint32) []byte {
	return order.AppendUint32(bs, v)
}

func WriteUint64(bs []byte, v uint64) []byte {
	return order.AppendUint64(bs, v)
}

// Signed Integers
func WriteInt8(bs []byte, v int8) []byte {
	return WriteUint8(bs, uint8(v))
}

func WriteInt16(bs []byte, v int16) []byte {
	return order.AppendUint16(bs, uint16(v))
}

func WriteInt32(bs []byte, v int32) []byte {
	return order.AppendUint32(bs, uint32(v))
}

func WriteInt64(bs []byte, v int64) []byte {
	return order.AppendUint64(bs, uint64(v))
}

//--------------------------------------------------------------------------------
// Read/Unappend values - Fixed Width
//--------------------------------------------------------------------------------

// Unsigned Integers
func ReadUint8(bs []byte) (uint8, int, error) {
	n := sizeUint8
	if len(bs) < n { return 0, 0, ErrTruncatedData }
	return bs[0], n, nil
}
func ReadUint16(bs []byte) (uint16, int, error) {
	n := sizeUint16
	if len(bs) < n { return 0, 0, ErrTruncatedData }
	return order.Uint16(bs), n, nil
}
func ReadUint32(bs []byte) (uint32, int, error) {
	n := sizeUint32
	if len(bs) < n { return 0, 0, ErrTruncatedData }
	return order.Uint32(bs), n, nil
}
func ReadUint64(bs []byte) (uint64, int, error) {
	n := sizeUint64
	if len(bs) < n { return 0, 0, ErrTruncatedData }
	return order.Uint64(bs), n, nil
}

// Signed Integers
func ReadInt8(bs []byte) (int8, int, error) {
	v, n, err := ReadUint8(bs)
	return int8(v), n, err
}
func ReadInt16(bs []byte) (int16, int, error) {
	v, n, err := ReadUint16(bs)
	return int16(v), n, err
}
func ReadInt32(bs []byte) (int32, int, error) {
	v, n, err := ReadUint32(bs)
	return int32(v), n, err
}
func ReadInt64(bs []byte) (int64, int, error) {
	v, n, err := ReadUint64(bs)
	return int64(v), n, err
}

//--------------------------------------------------------------------------------
// Read/Append - Variable Width
//--------------------------------------------------------------------------------
// Unsigned Integers
func WriteVarUint16(bs []byte, v uint16) []byte {
	return WriteVarUint64(bs, uint64(v))
}

func WriteVarUint32(bs []byte, v uint32) []byte {
	return WriteVarUint64(bs, uint64(v))
}

func WriteVarUint64(bs []byte, v uint64) []byte {
	return binary.AppendUvarint(bs, v)
}

// Signed Integers
func WriteVarInt16(bs []byte, v int16) []byte {
	return WriteVarInt64(bs, int64(v))
}

func WriteVarInt32(bs []byte, v int32) []byte {
	return WriteVarInt64(bs, int64(v))
}

func WriteVarInt64(bs []byte, v int64) []byte {
	return binary.AppendVarint(bs, v)
}

//--------------------------------------------------------------------------------
// Read/Unappend values - Variable Width
//--------------------------------------------------------------------------------

// Unsigned Integers

func ReadVarUint16(bs []byte) (uint16, int, error) {
	v, n, err := ReadVarUint64(bs)
	if err != nil { return 0, 0, err }
	return uint16(v), n, err
}
func ReadVarUint32(bs []byte) (uint32, int, error) {
	v, n, err := ReadVarUint64(bs)
	if err != nil { return 0, 0, err }
	return uint32(v), n, err
}
func ReadVarUint64(bs []byte) (uint64, int, error) {
	val, n := binary.Uvarint(bs)
	if n <= 0 { return 0, 0, ErrTruncatedData }
	return val, n, nil
}

// Signed Integers
func ReadVarInt16(bs []byte) (int16, int, error) {
	v, n, err := ReadVarInt64(bs)
	if err != nil { return 0, 0, err }
	return int16(v), n, err
}
func ReadVarInt32(bs []byte) (int32, int, error) {
	v, n, err := ReadVarInt64(bs)
	if err != nil { return 0, 0, err }
	return int32(v), n, err
}
func ReadVarInt64(bs []byte) (int64, int, error) {
	val, n := binary.Varint(bs)
	if n <= 0 { return 0, 0, ErrTruncatedData }
	return val, n, nil
}

//--------------------------------------------------------------------------------
// Complex types
//--------------------------------------------------------------------------------
func writeByteSlice(bs []byte, v []byte) []byte {
	bs = WriteVarUint64(bs, uint64(len(v)))
	bs = append(bs, v...)
	return bs
}

// This currently does not copy the data, hence it is private
func readByteSlice(bs []byte) ([]byte, int, error) {
	l, n, err := ReadVarUint64(bs)
	if err != nil { return nil, 0, err }

	startRead := n
	endRead := n + int(l)
	if len(bs) < endRead { return nil, 0, ErrTruncatedData }

	ret := bs[startRead:endRead]
	return ret, endRead, nil
}

func WriteString(bs []byte, v string) []byte {
	return writeByteSlice(bs, []byte(v))
}

func ReadString(bs []byte) (string, int, error) {
	dat, n, err := readByteSlice(bs)
	if err != nil { return "", 0, err }

	return string(dat), n, nil
}


func WriteBool(bs []byte, v bool) []byte {
	val := uint8(0)
	if v {
		val = 1
	}
	return WriteUint8(bs, val)
}
func ReadBool(bs []byte) (bool, int, error) {
	val, n, err := ReadUint8(bs)
	if err != nil { return false, 0, err }

	ret := false
	if val == 1 {
		ret = true
	}

	return ret, n, nil
}

// Floats
func WriteFloat32(bs []byte, v float32) []byte {
	return WriteUint32(bs, math.Float32bits(v))
}
func ReadFloat32(bs []byte) (float32, int, error) {
	v, n, err := ReadUint32(bs)
	if err != nil { return 0, 0, err }

	ret := math.Float32frombits(v)
	return ret, n, nil
}

func WriteFloat64(bs []byte, v float64) []byte {
	return WriteUint64(bs, math.Float64bits(v))
}
func ReadFloat64(bs []byte) (float64, int, error) {
	v, n, err := ReadUint64(bs)
	if err != nil { return 0, 0, err }

	ret := math.Float64frombits(v)
	return ret, n, nil
}
