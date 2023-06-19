package main

import (
	"github.com/unitoftime/cod/alt"
	"github.com/unitoftime/cod/alt/test"
	"github.com/unitoftime/cod/alt/inner"
)

func main() {
	var t1 inner.Type1
	var t2 inner.Type2
	var t4 inner.Type4
	var t3 inner.Type3
	alt.Generate(
		inner.AAA(5),
		t1,
		t2,
		t4,
		t3,
		test.Person{},
	)
}
