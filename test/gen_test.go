package main

import (
	"testing"

	"github.com/unitoftime/cod/test/subpackage"
)

func TestBlankStruct(t *testing.T) {
	d := BlankStruct{}

	res := BlankStruct{}

	bs := []byte{}
	bs = d.EncodeCod(bs)
	_, err := res.DecodeCod(bs)
	if err != nil { panic(err) }

	t.Log(d)
	t.Log(res)
}

func TestSubPackage(t *testing.T) {
	d := MyStruct{
		// Vector: subpackage.Vec{1, 2},

		Vector: []subpackage.Vec{
			subpackage.Vec{1, 2},
			subpackage.Vec{3, 4},
		},
	}

	res := MyStruct{}

	bs := []byte{}
	bs = d.EncodeCod(bs)
	_, err := res.DecodeCod(bs)
	if err != nil { panic(err) }

	t.Log(d)
	t.Log(res)
}

func TestPerson(t *testing.T) {
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
	}

	res := Person{
		// Map: make(map[string][]uint64),
		// MultiMap: make(map[string]map[uint32][]uint8),
	}

	// buf := cod.NewBuffer(0)
	bs := []byte{}
	bs = d.EncodeCod(bs)
	_, err := res.DecodeCod(bs)
	if err != nil { panic(err) }

	t.Log(d)
	t.Log(res)
}

func BenchmarkPerson(b *testing.B) {
	input := Person{
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
	}

	res := Person{
		// Map: make(map[string][]uint64),
		// MultiMap: make(map[string]map[uint32][]uint8),
	}

	// buf = cod.NewBuffer(1024)
	buf := make([]byte, 1024)

	var serialSize int
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		buf = input.EncodeCod(buf)

		_, err := res.DecodeCod(buf)
		if err != nil { panic(err) }

		serialSize += len(buf)
	}
	b.ReportMetric(float64(serialSize)/float64(b.N), "B/serial")
}
