package main

import (
	"github.com/unitoftime/cod"
	"github.com/unitoftime/cod/backend"
)

func (t Person) EncodeCod(bs []byte) []byte {

	bs = backend.WriteString(bs, t.Name)

	bs = backend.WriteUint8(bs, t.Age)

	bs = t.Id.EncodeCod(bs)
	for i0 := range t.Array {

		bs = backend.WriteVarUint16(bs, t.Array[i0])

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Slice)))
		for i0 := range t.Slice {

			bs = backend.WriteVarUint32(bs, t.Slice[i0])

		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice)))
		for i0 := range t.DoubleSlice {

			{
				bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice[i0])))
				for i1 := range t.DoubleSlice[i0] {

					bs = backend.WriteUint8(bs, t.DoubleSlice[i0][i1])

				}
			}
		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Map)))

		for k0, v0 := range t.Map {

			bs = backend.WriteString(bs, k0)

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v0)))
				for i1 := range v0 {

					bs = backend.WriteVarUint64(bs, v0[i1])

				}
			}
		}

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.MultiMap)))

		for k0, v0 := range t.MultiMap {

			bs = backend.WriteString(bs, k0)

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v0)))

				for k1, v1 := range v0 {

					bs = backend.WriteVarUint32(bs, k1)

					{
						bs = backend.WriteVarUint64(bs, uint64(len(v1)))
						for i2 := range v1 {

							bs = backend.WriteUint8(bs, v1[i2])

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

	nOff, err = t.Id.DecodeCod(bs[n:])
	n += nOff

	for i0 := range t.Array {

		t.Array[i0], nOff, err = backend.ReadVarUint16(bs[n:])
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
		length, nOff, err := backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		for i0 := 0; i0 < int(length); i0++ {
			var value0 uint32

			value0, nOff, err = backend.ReadVarUint32(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			if err != nil {
				return 0, err
			}

			t.Slice = append(t.Slice, value0)
		}
	}
	{
		var length uint64
		length, nOff, err := backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		for i0 := 0; i0 < int(length); i0++ {
			var value0 []uint8

			{
				var length uint64
				length, nOff, err := backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				for i1 := 0; i1 < int(length); i1++ {
					var value1 uint8

					value1, nOff, err = backend.ReadUint8(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					if err != nil {
						return 0, err
					}

					value0 = append(value0, value1)
				}
			}
			if err != nil {
				return 0, err
			}

			t.DoubleSlice = append(t.DoubleSlice, value0)
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

		for i0 := 0; i0 < int(length); i0++ {
			var key0 string
			var val0 []uint64

			key0, nOff, err = backend.ReadString(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff

			{
				var length uint64
				length, nOff, err := backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				for i1 := 0; i1 < int(length); i1++ {
					var value1 uint64

					value1, nOff, err = backend.ReadVarUint64(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					if err != nil {
						return 0, err
					}

					val0 = append(val0, value1)
				}
			}
			if err != nil {
				return 0, err
			}

			t.Map[key0] = val0
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

		for i0 := 0; i0 < int(length); i0++ {
			var key0 string
			var val0 map[uint32][]uint8

			key0, nOff, err = backend.ReadString(bs[n:])
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

				if val0 == nil {
					val0 = make(map[uint32][]uint8)
				}

				for i1 := 0; i1 < int(length); i1++ {
					var key1 uint32
					var val1 []uint8

					key1, nOff, err = backend.ReadVarUint32(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					{
						var length uint64
						length, nOff, err := backend.ReadVarUint64(bs[n:])
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

					val0[key1] = val1
				}
			}
			if err != nil {
				return 0, err
			}

			t.MultiMap[key0] = val0
		}
	}
	nOff, err = t.MyUnion.DecodeCod(bs[n:])
	n += nOff

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

	default:
		panic("unknown type placed in union")
	}
	return n, err

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
					length, nOff, err := backend.ReadVarUint64(bs[n:])
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
