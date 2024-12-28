package test

import (
	"github.com/unitoftime/cod/backend"

	"github.com/unitoftime/cod/test/subpackage/blocked"

	"github.com/unitoftime/cod"

	"github.com/unitoftime/ecs"

	"github.com/unitoftime/cod/test/subpackage"
)

func (t BlankStruct) EncodeCod(bs []byte) []byte {
	return bs
}

func (t *BlankStruct) DecodeCod(bs []byte) (n int, err error) {
	return
}

func (t BlankStruct) CodEquals(tt BlankStruct) bool {
	return true
}

func (t BlockedStruct) EncodeCod(bs []byte) []byte {

	bs = backend.WriteVarUint64(bs, uint64(t.Basic))

	return bs
}

func (t *BlockedStruct) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var decoded uint64
		decoded, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Basic = blocked.Basic(decoded)
	}

	// println("BlockedStruct:", n)
	return n, err
}

func (t BlockedStruct) CodEquals(tt BlockedStruct) bool {

	if t.Basic != tt.Basic {
		return false
	}

	return true
}

func (t BlockedStruct2) EncodeCod(bs []byte) []byte {

	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Basic)))
		for i1 := range t.Basic {

			bs = backend.WriteVarUint64(bs, uint64(t.Basic[i1]))

		}
	}
	return bs
}

func (t *BlockedStruct2) DecodeCod(bs []byte) (int, error) {
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
			var value1 blocked.Basic

			{
				var decoded uint64
				decoded, nOff, err = backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				value1 = blocked.Basic(decoded)
			}

			if err != nil {
				return 0, err
			}

			t.Basic = append(t.Basic, value1)
		}
	}

	// println("BlockedStruct2:", n)
	return n, err
}

func (t BlockedStruct2) CodEquals(tt BlockedStruct2) bool {

	{
		if len(t.Basic) != len(tt.Basic) {
			return false
		}
		for i1 := range t.Basic {

			if t.Basic[i1] != tt.Basic[i1] {
				return false
			}

		}
	}
	return true
}

func (t Id) EncodeCod(bs []byte) []byte {

	bs = backend.WriteVarUint16(bs, (t.Val))

	return bs
}

func (t *Id) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var decoded uint16
		decoded, nOff, err = backend.ReadVarUint16(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Val = (decoded)
	}

	// println("Id:", n)
	return n, err
}

func (t Id) CodEquals(tt Id) bool {

	if t.Val != tt.Val {
		return false
	}

	return true
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

	// println("MyStruct:", n)
	return n, err
}

func (t MyStruct) CodEquals(tt MyStruct) bool {

	{
		if len(t.Vector) != len(tt.Vector) {
			return false
		}
		for i1 := range t.Vector {

			if !t.Vector[i1].CodEquals(tt.Vector[i1]) {
				return false
			}

		}
	}
	return true
}

func (t MyUnion) EncodeCod(bs []byte) []byte {

	tag := t.Tag()
	bs = backend.WriteUint8(bs, tag)
	if tag == 0 {
		// Zero tag indicates nil, so write nothing else
		return bs
	}

	rawVal := t.Get()
	bs = rawVal.EncodeCod(bs)

	return bs

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
		return 0, backend.ErrUnknownUnionType
	}

	// println("MyUnion:", n)
	return n, err
}

func (t MyUnion) Tag() uint8 {
	rawVal := t.Get()
	if rawVal == nil {
		// Zero tag indicates nil
		return 0
	}

	switch rawVal.(type) {

	case Id:
		return 1

	case SpecialMap:
		return 2

	case subpackage.Vec:
		return 3

	default:
		panic("unknown type placed in union")
	}
}

func (t MyUnion) Size() int {
	return 4
}

func (t MyUnion) CodEquals(tt MyUnion) bool {
	if t.Tag() != tt.Tag() {
		return false
	}

	rawVal := t.Get()
	switch sv := rawVal.(type) {

	case Id:
		sv2 := tt.Get().(Id)
		return sv.CodEquals(sv2)

	case SpecialMap:
		sv2 := tt.Get().(SpecialMap)
		return sv.CodEquals(sv2)

	case subpackage.Vec:
		sv2 := tt.Get().(subpackage.Vec)
		return sv.CodEquals(sv2)

	default:
		panic("unknown type placed in union")
	}

	return true
}

func (t Person) EncodeCod(bs []byte) []byte {

	bs = backend.WriteString(bs, (t.Name))

	bs = backend.WriteUint8(bs, (t.Age))

	bs = t.Id.EncodeCod(bs)
	for i1 := range t.Array {

		bs = backend.WriteVarUint16(bs, (t.Array[i1]))

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Slice)))
		for i1 := range t.Slice {

			bs = backend.WriteVarUint32(bs, (t.Slice[i1]))

		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice)))
		for i1 := range t.DoubleSlice {

			{
				bs = backend.WriteVarUint64(bs, uint64(len(t.DoubleSlice[i1])))
				for i2 := range t.DoubleSlice[i1] {

					bs = backend.WriteUint8(bs, (t.DoubleSlice[i1][i2]))

				}
			}
		}
	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.Map)))

		for k1, v1 := range t.Map {

			bs = backend.WriteString(bs, (k1))

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v1)))
				for i2 := range v1 {

					bs = backend.WriteVarUint64(bs, (v1[i2]))

				}
			}
		}

	}
	{
		bs = backend.WriteVarUint64(bs, uint64(len(t.MultiMap)))

		for k1, v1 := range t.MultiMap {

			bs = backend.WriteString(bs, (k1))

			{
				bs = backend.WriteVarUint64(bs, uint64(len(v1)))

				for k2, v2 := range v1 {

					bs = backend.WriteVarUint32(bs, (k2))

					{
						bs = backend.WriteVarUint64(bs, uint64(len(v2)))
						for i3 := range v2 {

							bs = backend.WriteUint8(bs, (v2[i3]))

						}
					}
				}

			}
		}

	}
	bs = t.MyUnion.EncodeCod(bs)
	{
		if t.Pointer == nil {
			// Zero tag indicates nil
			bs = backend.WriteUint8(bs, 0)
		} else {
			bs = backend.WriteUint8(bs, 1)
			value1 := *t.Pointer

			bs = value1.EncodeCod(bs)
		}
	}
	return bs
}

func (t *Person) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var decoded string
		decoded, nOff, err = backend.ReadString(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Name = (decoded)
	}

	{
		var decoded uint8
		decoded, nOff, err = backend.ReadUint8(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Age = (decoded)
	}

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

		{
			var decoded uint16
			decoded, nOff, err = backend.ReadVarUint16(bs[n:])
			if err != nil {
				return 0, err
			}
			n += nOff
			t.Array[i1] = (decoded)
		}

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

			{
				var decoded uint32
				decoded, nOff, err = backend.ReadVarUint32(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				value1 = (decoded)
			}

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

					{
						var decoded uint8
						decoded, nOff, err = backend.ReadUint8(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff
						value2 = (decoded)
					}

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

			{
				var decoded string
				decoded, nOff, err = backend.ReadString(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				key1 = (decoded)
			}

			{
				var length uint64
				length, nOff, err = backend.ReadVarUint64(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff

				for i2 := 0; i2 < int(length); i2++ {
					var value2 uint64

					{
						var decoded uint64
						decoded, nOff, err = backend.ReadVarUint64(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff
						value2 = (decoded)
					}

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

			{
				var decoded string
				decoded, nOff, err = backend.ReadString(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				key1 = (decoded)
			}

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

					{
						var decoded uint32
						decoded, nOff, err = backend.ReadVarUint32(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff
						key2 = (decoded)
					}

					{
						var length uint64
						length, nOff, err = backend.ReadVarUint64(bs[n:])
						if err != nil {
							return 0, err
						}
						n += nOff

						for i3 := 0; i3 < int(length); i3++ {
							var value3 uint8

							{
								var decoded uint8
								decoded, nOff, err = backend.ReadUint8(bs[n:])
								if err != nil {
									return 0, err
								}
								n += nOff
								value3 = (decoded)
							}

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

	{
		var tagVal uint8
		tagVal, nOff, err = backend.ReadUint8(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		if tagVal == 0 {
			// Zero tag indicates nil
			t.Pointer = nil
		} else {
			var value1 BlockedStruct

			{
				var decoded BlockedStruct
				nOff, err = decoded.DecodeCod(bs[n:])
				if err != nil {
					return 0, err
				}
				n += nOff
				value1 = decoded
			}

			t.Pointer = &value1
		}
	}

	// println("Person:", n)
	return n, err
}

func (t Person) CodEquals(tt Person) bool {

	if t.Name != tt.Name {
		return false
	}

	if t.Age != tt.Age {
		return false
	}

	if !t.Id.CodEquals(tt.Id) {
		return false
	}

	for i1 := range t.Array {

		if t.Array[i1] != tt.Array[i1] {
			return false
		}

	}
	{
		if len(t.Slice) != len(tt.Slice) {
			return false
		}
		for i1 := range t.Slice {

			if t.Slice[i1] != tt.Slice[i1] {
				return false
			}

		}
	}
	{
		if len(t.DoubleSlice) != len(tt.DoubleSlice) {
			return false
		}
		for i1 := range t.DoubleSlice {

			{
				if len(t.DoubleSlice[i1]) != len(tt.DoubleSlice[i1]) {
					return false
				}
				for i2 := range t.DoubleSlice[i1] {

					if t.DoubleSlice[i1][i2] != tt.DoubleSlice[i1][i2] {
						return false
					}

				}
			}
		}
	}
	{
		if len(t.Map) != len(tt.Map) {
			return false
		}
		for k1, v1 := range t.Map {
			tv1, ok := tt.Map[k1]
			if !ok {
				return false
			}

			{
				if len(v1) != len(tv1) {
					return false
				}
				for i2 := range v1 {

					if v1[i2] != tv1[i2] {
						return false
					}

				}
			}
		}
	}
	{
		if len(t.MultiMap) != len(tt.MultiMap) {
			return false
		}
		for k1, v1 := range t.MultiMap {
			tv1, ok := tt.MultiMap[k1]
			if !ok {
				return false
			}

			{
				if len(v1) != len(tv1) {
					return false
				}
				for k2, v2 := range v1 {
					tv2, ok := tv1[k2]
					if !ok {
						return false
					}

					{
						if len(v2) != len(tv2) {
							return false
						}
						for i3 := range v2 {

							if v2[i3] != tv2[i3] {
								return false
							}

						}
					}
				}
			}
		}
	}
	if !t.MyUnion.CodEquals(tt.MyUnion) {
		return false
	}

	{
		tNil := (t.Pointer == nil)
		ttNil := (tt.Pointer == nil)
		if tNil != ttNil {
			return false
		}
		if !tNil && !ttNil {
			value1 := *t.Pointer
			tvalue1 := *tt.Pointer

			if !value1.CodEquals(tvalue1) {
				return false
			}

		}
	}
	return true
}

func (t SpecialMap) EncodeCod(bs []byte) []byte {

	{
		value0 := map[string][]uint8(t)

		{
			bs = backend.WriteVarUint64(bs, uint64(len(value0)))

			for k1, v1 := range value0 {

				bs = backend.WriteString(bs, (k1))

				{
					bs = backend.WriteVarUint64(bs, uint64(len(v1)))
					for i2 := range v1 {

						bs = backend.WriteUint8(bs, (v1[i2]))

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

				{
					var decoded string
					decoded, nOff, err = backend.ReadString(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff
					key1 = (decoded)
				}

				{
					var length uint64
					length, nOff, err = backend.ReadVarUint64(bs[n:])
					if err != nil {
						return 0, err
					}
					n += nOff

					for i2 := 0; i2 < int(length); i2++ {
						var value2 uint8

						{
							var decoded uint8
							decoded, nOff, err = backend.ReadUint8(bs[n:])
							if err != nil {
								return 0, err
							}
							n += nOff
							value2 = (decoded)
						}

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

	// println("SpecialMap:", n)
	return n, err
}

func (t SpecialMap) CodEquals(tt SpecialMap) bool {

	{
		value0 := map[string][]uint8(t)
		tvalue0 := map[string][]uint8(tt)

		{
			if len(value0) != len(tvalue0) {
				return false
			}
			for k1, v1 := range value0 {
				tv1, ok := tvalue0[k1]
				if !ok {
					return false
				}

				{
					if len(v1) != len(tv1) {
						return false
					}
					for i2 := range v1 {

						if v1[i2] != tv1[i2] {
							return false
						}

					}
				}
			}
		}
	}
	return true
}

func (t MyUnion) Get() cod.EncoderDecoder {
	codUnion := cod.Union(t)
	rawVal := codUnion.GetRawValue()
	return rawVal

	// switch rawVal.(type) {
	// <no value>
	// default:
	//    panic("unknown type placed in union")
	// }
}

func (t *MyUnion) Set(v cod.EncoderDecoder) {
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

func NewMyUnion(v cod.EncoderDecoder) MyUnion {
	var ret MyUnion
	ret.Set(v)
	return ret
}

var BlockedStructComp = ecs.Comp(BlockedStruct{})

func (c BlockedStruct) CompId() ecs.CompId {
	return BlockedStructComp.CompId()
}
func (c BlockedStruct) CompWrite(w ecs.W) {
	BlockedStructComp.WriteVal(w, c)
}
