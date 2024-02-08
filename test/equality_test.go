package test

import (
	"fmt"
	"testing"

	"github.com/unitoftime/cod"
	"github.com/unitoftime/cod/test/subpackage"
	"github.com/unitoftime/cod/test/subpackage/blocked"
)

func serializedCompare[A cod.EncoderDecoder](a A, b A) bool {
	a_bs := []byte{}
	a_bs = a.EncodeCod(a_bs)

	b_bs := []byte{}
	b_bs = b.EncodeCod(b_bs)

	if len(a_bs) != len(b_bs) {
		return false
	}
	for i := range a_bs {
		if a_bs[i] != b_bs[i] {
			return false
		}
	}

	return true
}

func generatedCompare[A interface{CodEquals(A) bool}](a A, b A) bool {
	return a.CodEquals(b)
}

func TestBlockedStructEquality(t *testing.T) {
	d := BlockedStruct{
		Basic: blocked.Basic(1),
	}

	if !serializedCompare(d, d) {
		t.Error("SERIALZE MISMATCH")
	}

	if !generatedCompare(d, d) {
		t.Error("MISMATCH")
	}
}

func TestBlockedStructNotEqual(t *testing.T) {
	d := BlockedStruct{
		Basic: blocked.Basic(1),
	}
	d2 := BlockedStruct{
		Basic: blocked.Basic(2),
	}

	if serializedCompare(d, d2) {
		t.Error("SER: MATCH-BUT THEY SHOULDNT")
	}

	if generatedCompare(d, d2) {
		t.Error("MATCH-BUT THEY SHOULDNT")
	}
}

func TestBlockedStruct2Equality(t *testing.T) {
	d := BlockedStruct2{
		Basic: []blocked.Basic{
			blocked.Basic(1),
			blocked.Basic(2),
		},
	}

	if !serializedCompare(d, d) {
		t.Error("SERIALZE MISMATCH")
	}

	if !generatedCompare(d, d) {
		t.Error("MISMATCH")
	}

	// t.Log(d)
	// t.Log(res)
}


func TestBlankStructEquality(t *testing.T) {
	d := BlankStruct{}

	if !serializedCompare(d, d) {
		t.Error("SERIALZE MISMATCH")
	}

	if !generatedCompare(d, d) {
		t.Error("MISMATCH")
	}
}

func TestSubPackageEquality(t *testing.T) {
	d := MyStruct{
		// Vector: subpackage.Vec{1, 2},

		Vector: []subpackage.Vec{
			subpackage.Vec{1, 2},
			subpackage.Vec{3, 4},
		},
	}

	if !serializedCompare(d, d) {
		t.Error("SERIALZE MISMATCH")
	}

	if !generatedCompare(d, d) {
		t.Error("MISMATCH")
	}
}

func TestPersonEquality(t *testing.T) {
	d := Person{
		Name: "hello",
		Age: 5,
		Id: Id{7},
		Array: [2]uint16{8, 9},
		Slice: []uint32{100, 101, 102},
		DoubleSlice: [][]uint8{[]uint8{1, 2, 3}, []uint8{4, 5, 6}},

		Map: map[string][]uint64{
			"a": []uint64{1000, 2000, 3000},
			"b": []uint64{4000, 5000, 6000},
		},
		MultiMap: map[string]map[uint32][]uint8{
			"c": map[uint32][]uint8{
				1: []uint8{11, 12},
				2: []uint8{22, 23},
			},
			"d": map[uint32][]uint8{
				3: []uint8{33, 34},
				4: []uint8{44, 45},
			},
		},

		MyUnion: NewMyUnion(Id{8}),

		Pointer: &BlockedStruct{
			Basic: blocked.Basic(1),
		},
	}

	// if !serializedCompare(d, d) {
	// 	t.Error("SERIALZE MISMATCH")
	// }

	if !generatedCompare(d, d) {
		t.Error("MISMATCH")
	}
}

func TestPersonEquality2(t *testing.T) {
	d := Person{
		Name: "hello",
		Age: 5,
		Id: Id{7},
		Array: [2]uint16{8, 9},
		Slice: []uint32{100, 101, 102},
		DoubleSlice: [][]uint8{[]uint8{1, 2, 3}, []uint8{4, 5, 6}},

		Map: map[string][]uint64{
			"a": []uint64{1000, 2000, 3000},
			"b": []uint64{4000, 5000, 6000},
		},
		MultiMap: map[string]map[uint32][]uint8{
			"c": map[uint32][]uint8{
				1: []uint8{11, 12},
				2: []uint8{22, 23},
			},
			"d": map[uint32][]uint8{
				3: []uint8{33, 34},
				4: []uint8{44, 45},
			},
		},

		MyUnion: NewMyUnion(Id{8}),

		Pointer: &BlockedStruct{
			Basic: blocked.Basic(1),
		},
	}

	d2 := Person{
		Name: "hello",
		Age: 5,
		Id: Id{7},
		Array: [2]uint16{8, 9},
		Slice: []uint32{100, 101, 102},
		DoubleSlice: [][]uint8{[]uint8{1, 2, 3}, []uint8{4, 5, 6}},

		Map: map[string][]uint64{
			"a": []uint64{1000, 2000, 3000},
			"b": []uint64{4000, 5000, 6000},
		},
		MultiMap: map[string]map[uint32][]uint8{
			"c": map[uint32][]uint8{
				1: []uint8{11, 12},
				2: []uint8{22, 23},
			},
			"d": map[uint32][]uint8{
				3: []uint8{33, 34},
				4: []uint8{44, 45},
			},
		},

		MyUnion: NewMyUnion(Id{8}),

		Pointer: &BlockedStruct{
			Basic: blocked.Basic(1),
		},
	}

	// if !serializedCompare(d, d2) {
	// 	t.Error("SERIALZE MISMATCH")
	// }

	if !generatedCompare(d, d2) {
		t.Error("MISMATCH")
	}
}

func TestPersonEqualityNotEqual(t *testing.T) {
	d := Person{
		Name: "hello",
		Age: 5,
		Id: Id{7},
		Array: [2]uint16{8, 9},
		Slice: []uint32{100, 101, 102},
		DoubleSlice: [][]uint8{[]uint8{1, 2, 3}, []uint8{4, 5, 6}},

		Map: map[string][]uint64{
			"a": []uint64{1000, 2000, 3000},
			"b": []uint64{4000, 5000, 6000},
		},
		MultiMap: map[string]map[uint32][]uint8{
			"c": map[uint32][]uint8{
				1: []uint8{11, 12},
				2: []uint8{22, 23},
			},
			"d": map[uint32][]uint8{
				3: []uint8{33, 34},
				4: []uint8{44, 45},
			},
		},

		MyUnion: NewMyUnion(Id{8}),

		Pointer: &BlockedStruct{
			Basic: blocked.Basic(1),
		},
	}

	d2 := Person{
		Name: "hello",
		Age: 5,
		Id: Id{7},
		Array: [2]uint16{8, 9},
		Slice: []uint32{100, 101, 102},
		DoubleSlice: [][]uint8{[]uint8{1, 2, 3}, []uint8{4, 5, 6}},

		Map: map[string][]uint64{
			"a": []uint64{1000, 2000, 3000},
			"b": []uint64{4000, 5000, 6000},
		},
		MultiMap: map[string]map[uint32][]uint8{
			"c": map[uint32][]uint8{
				1: []uint8{11, 12},
				2: []uint8{22, 23},
			},
			"d": map[uint32][]uint8{
				3: []uint8{33, 34},
				4: []uint8{44, 45},
			},
		},

		MyUnion: NewMyUnion(Id{9}),//<------ you changed this part

		Pointer: &BlockedStruct{
			Basic: blocked.Basic(1),
		},
	}

	// if !serializedCompare(d, d2) {
	// 	t.Error("SERIALZE MISMATCH")
	// }


	fmt.Println("----")
	fmt.Println(generatedCompare(d, d2))
	fmt.Println("----")
	if generatedCompare(d, d2) {
		t.Error("SHOULD HAVE MISMATCHED")
	}
}
