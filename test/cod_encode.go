package main

import (
	"github.com/unitoftime/cod/backend"

	"github.com/unitoftime/cod"

	"github.com/unitoftime/cod/test/subpackage"
)

func (t Person) EncodeCod(bs []byte) []byte {

	bs = backend.WriteString(bs, t.Name)

	bs = backend.WriteUint8(bs, t.Age)

	bs = t.Id.EncodeCod(bs)
	for i1 := range t.Array {

		bs = backend.WriteVarUint16(bs, t.Array[i1])

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Slice)))
		for i1 := range t.Slice {

			bs = backend.WriteVarUint32(bs, t.Slice[i1])

		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice)))
		for i1 := range t.DoubleSlice {

			{
				bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice[i1])))
				for i2 := range t.DoubleSlice[i1] {

					bs = backend.WriteUint8(bs, t.DoubleSlice[i1][i2])

				}
			}
		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Map)))

		for k1, v1 := range t.Map {

			bs = backend.WriteString(bs, k1)

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v1)))
				for i2 := range v1 {

					bs = backend.WriteVarUint64(bs, v1[i2])

				}
			}
		}

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.MultiMap)))

		for k1, v1 := range t.MultiMap {

			bs = backend.WriteString(bs, k1)

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v1)))

				for k2, v2 := range v1 {

					bs = backend.WriteVarUint32(bs, k2)

					{
						bs = backend.WriteVarUint64(bs, uint64(len(v2)))
						for i3 := range v2 {

							bs = backend.WriteUint8(bs, v2[i3])

						}
					}
				}

			}
		}

	}
	bs = t.MyUnion.EncodeCod(bs)
	return bs
}

func (t *Person) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	t.Name, nOff, err = backend.ReadString(bs[n:])
	if err != nil {
		return 0, err
	}
	n += nOff

	t.Age, nOff, err = backend.ReadUint8(bs[n:])
	if err != nil {
		return 0, err
	}
	n += nOff

	{
		var decoded Id
		nOff, err = decoded.DecodeCod(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Id = decoded
	}

	for i1 := range t.Array {

		t.Array[i1], nOff, err = backend.ReadVarUint16(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		if err != nil {
			return 0, err
		}
	}
	{
		var length uint64
		length, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		for i1 := 0; i1 < int(length); i1++ {
			var value1 uint32

			value1, nOff, err = backend.ReadVarUint32(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			if err != nil {
				return 0, err
			}

			t.Slice = append(t.Slice, value1)
		}
	}
	{
		var length uint64
		length, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		for i1 := 0; i1 < int(length); i1++ {
			var value1 []uint8

			{
				var length uint64
				length, nOff, err = backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				for i2 := 0; i2 < int(length); i2++ {
					var value2 uint8

					value2, nOff, err = backend.ReadUint8(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					if err != nil {
						return 0, err
					}

					value1 = append(value1, value2)
				}
			}
			if err != nil {
				return 0, err
			}

			t.DoubleSlice = append(t.DoubleSlice, value1)
		}
	}
	{
		var length uint64
		length, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		if t.Map == nil {
			t.Map = make(map[string][]uint64)
		}

		for i1 := 0; i1 < int(length); i1++ {
			var key1 string
			var val1 []uint64

			key1, nOff, err = backend.ReadString(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			{
				var length uint64
				length, nOff, err = backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				for i2 := 0; i2 < int(length); i2++ {
					var value2 uint64

					value2, nOff, err = backend.ReadVarUint64(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					if err != nil {
						return 0, err
					}

					val1 = append(val1, value2)
				}
			}
			if err != nil {
				return 0, err
			}

			t.Map[key1] = val1
		}
	}
	{
		var length uint64
		length, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		if t.MultiMap == nil {
			t.MultiMap = make(map[string]map[uint32][]uint8)
		}

		for i1 := 0; i1 < int(length); i1++ {
			var key1 string
			var val1 map[uint32][]uint8

			key1, nOff, err = backend.ReadString(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			{
				var length uint64
				length, nOff, err = backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				if val1 == nil {
					val1 = make(map[uint32][]uint8)
				}

				for i2 := 0; i2 < int(length); i2++ {
					var key2 uint32
					var val2 []uint8

					key2, nOff, err = backend.ReadVarUint32(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					{
						var length uint64
						length, nOff, err = backend.ReadVarUint64(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff

						for i3 := 0; i3 < int(length); i3++ {
							var value3 uint8

							value3, nOff, err = backend.ReadUint8(bs[n:])
							if err != nil {
								return 0, err
							}
							n += nOff

							if err != nil {
								return 0, err
							}

							val2 = append(val2, value3)
						}
					}
					if err != nil {
						return 0, err
					}

					val1[key2] = val2
				}
			}
			if err != nil {
				return 0, err
			}

			t.MultiMap[key1] = val1
		}
	}
	{
		var decoded MyUnion
		nOff, err = decoded.DecodeCod(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.MyUnion = decoded
	}

	return n, err
}

func (t MyUnion) EncodeCod(bs []byte) []byte {

	rawVal := t.Get()
	if rawVal == nil {
		// Zero tag indicates nil
		bs = backend.WriteUint8(bs, 0)
		return bs
	}

	switch sv := rawVal.(type) {

	case Id:
		bs = backend.WriteUint8(bs, 1)
		bs = sv.EncodeCod(bs)

	case SpecialMap:
		bs = backend.WriteUint8(bs, 2)
		bs = sv.EncodeCod(bs)

	case subpackage.Vec:
		bs = backend.WriteUint8(bs, 3)
		bs = sv.EncodeCod(bs)

	default:
		panic("unknown type placed in union")
	}

	return bs
}

func (t *MyUnion) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	var tagVal uint8

	tagVal, nOff, err = backend.ReadUint8(bs[n:])
	if err != nil {
		return 0, err
	}
	n += nOff

	switch tagVal {
	case 0: // Zero tag indicates nil
		return 0, nil

	case 1:
		var decoded Id
		nOff, err = decoded.DecodeCod(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		t.Set(decoded)

	case 2:
		var decoded SpecialMap
		nOff, err = decoded.DecodeCod(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		t.Set(decoded)

	case 3:
		var decoded subpackage.Vec
		nOff, err = decoded.DecodeCod(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		t.Set(decoded)

	default:
		panic("unknown type placed in union")
	}

	return n, err
}

func (t SpecialMap) EncodeCod(bs []byte) []byte {

	{
		value0 := map[string][]uint8(t)

		{
			bs = backend.WriteVarUint64(bs, uint64(len(value0)))

			for k1, v1 := range value0 {

				bs = backend.WriteString(bs, k1)

				{
					bs = backend.WriteVarUint64(bs, uint64(len(v1)))
					for i2 := range v1 {

						bs = backend.WriteUint8(bs, v1[i2])

					}
				}
			}

		}

	}
	return bs
}

func (t *SpecialMap) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var value0 map[string][]uint8

		{
			var length uint64
			length, nOff, err = backend.ReadVarUint64(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			if value0 == nil {
				value0 = make(map[string][]uint8)
			}

			for i1 := 0; i1 < int(length); i1++ {
				var key1 string
				var val1 []uint8

				key1, nOff, err = backend.ReadString(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				{
					var length uint64
					length, nOff, err = backend.ReadVarUint64(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					for i2 := 0; i2 < int(length); i2++ {
						var value2 uint8

						value2, nOff, err = backend.ReadUint8(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff

						if err != nil {
							return 0, err
						}

						val1 = append(val1, value2)
					}
				}
				if err != nil {
					return 0, err
				}

				value0[key1] = val1
			}
		}
		*t = SpecialMap(value0)
	}

	return n, err
}

func (t Id) EncodeCod(bs []byte) []byte {

	bs = backend.WriteVarUint16(bs, t.Val)

	return bs
}

func (t *Id) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	t.Val, nOff, err = backend.ReadVarUint16(bs[n:])
	if err != nil {
		return 0, err
	}
	n += nOff

	return n, err
}

func (t MyStruct) EncodeCod(bs []byte) []byte {

	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Vector)))
		for i1 := range t.Vector {

			bs = t.Vector[i1].EncodeCod(bs)
		}
	}
	return bs
}

func (t *MyStruct) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var length uint64
		length, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		for i1 := 0; i1 < int(length); i1++ {
			var value1 subpackage.Vec

			{
				var decoded subpackage.Vec
				nOff, err = decoded.DecodeCod(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				value1 = decoded
			}

			if err != nil {
				return 0, err
			}

			t.Vector = append(t.Vector, value1)
		}
	}

	return n, err
}

func (t MyUnion) Get() any {
	codUnion := cod.Union(t)
	rawVal := codUnion.GetRawValue()
	return rawVal

	// switch rawVal.(type) {
	// <no value>
	// default:
	//    panic("unknown type placed in union")
	// }
}

func (t *MyUnion) Set(v any) {
	codUnion := cod.Union(*t)
	codUnion.PutRawValue(v)
	*t = MyUnion(codUnion)

	// switch tagVal {
	// case 0: // Zero tag indicates nil
	//    return nil

	// <no value>
	// default:
	//    panic("unknown type placed in union")
	// }
	// return err
}

func NewMyUnion(v any) MyUnion {
	var ret MyUnion
	ret.Set(v)
	return ret
}
