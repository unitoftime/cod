package test

import (
	"github.com/unitoftime/cod/test/subpackage"
)

//cod:struct
type MyStruct struct {
	Vector []subpackage.Vec
	// Vector subpackage.Vec
}
