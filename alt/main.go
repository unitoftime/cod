package main

import (
	"fmt"
	"reflect"
	"bytes"

	// "github.com/unitoftime/flow/phy2"
	"github.com/unitoftime/cod/alt/inner"
)

var supportedApis = map[string]string{
	"uint8": "Uint8",
	"int8": "Int8",

	"uint": "Uint",
	"int": "Int",

	// As a default, we always use Variable length encoding APIs for anything > 2 bytes
	"uint16": "VarUint16",
	"uint32": "VarUint32",
	"uint64": "VarUint64",

	"int16": "VarInt16",
	"int32": "VarInt32",
	"int64": "VarInt64",

	"float32": "Float32",
	"float64": "Float64",

	"string": "String",
	"bool": "Bool",
}
func isStandard(name string) bool {
	if name == "" { return true } // Anything without a name isnt type defined
	_, std := supportedApis[name]
	return std
}

func main() {
	gen := newGenerator()
	var t1 inner.Type1
	gen.search(reflect.TypeOf(t1))
	var t2 inner.Type2
	gen.search(reflect.TypeOf(t2))
	var t4 inner.Type4
	gen.search(reflect.TypeOf(t4))
	var t3 inner.Type3
	gen.search(reflect.TypeOf(t3))


	fmt.Println("-----")
	fmt.Println(gen.Types)
	fmt.Println("-----")

	buf := new(bytes.Buffer)
	gen.generate(buf)
	fmt.Println(buf.String())
}

type generator struct {
	Types map[reflect.Type]bool
}
func newGenerator() *generator {
	return &generator{
		Types: make(map[reflect.Type]bool),
	}
}

func (g *generator) addNewType(t reflect.Type) bool {
	if isStandard(t.Name()) { return true }

	_, exists := g.Types[t]
	if exists { return false }
	g.Types[t] = true
	return true
}

func (g *generator) search(t reflect.Type) {

	success := g.addNewType(t)
	if !success {
		fmt.Println("Type already exists. skipping")
	}
	fmt.Println("Details: ", t.Name(), t.PkgPath(), t.String(), t.Kind())


	switch t.Kind() {
	case reflect.Struct:
		numField := t.NumField()
		fmt.Println("genStruct: ", numField)
		for i := 0; i < numField; i++ {
			field := t.Field(i)
			fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Type.Kind())
			g.search(field.Type)
		}
	default:
		fmt.Println("unhandled kind: ", t.Kind())
	}
}

func (g *generator) generate(buf *bytes.Buffer) {
	for t := range g.Types {
		g.marshalFunc(buf, t)
		g.unmarshalFunc(buf, t)
	}
}


func (g *generator) marshalFunc(buf *bytes.Buffer, t reflect.Type) {
	inner := new(bytes.Buffer)

	if t.Kind() = reflect.Struct {
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			field := t.Field(i)
			marshalField(buf, field)
		}
	}

	err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
		"Type": t.Name(),
		"InnerCode": inner.String(),
	})
	if err != nil { panic(err) }
}

func (g *generator) unmarshalFunc(buf *bytes.Buffer, t reflect.Type) {
	err := BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
		"Type": t.Name(),
		"InnerCode": "TODO",
	})
	if err != nil { panic(err) }
}

func (g *generator) marshalField(buf *bytes.Buffer, t reflect.Type) {
	fmt.Println("marshalField: ", t.Name(), t.Kind())

	
	if t.Kind() = reflect.Struct {
		
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
				"Type": t.Name(),
				"InnerCode": "TODO",
			})
			if err != nil { panic(err) }
		}
	}

}

// func (g *generator) unmarshalFields(buf *bytes.Buffer, t reflect.Type) {
// 	err := BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
// 		"Type": t.Name(),
// 		"InnerCode": "TODO",
// 	})
// 	if err != nil { panic(err) }
// }

// func generateAny(a any) {

// 	fmt.Println("")
// 	fmt.Printf("generate: %T\n", a)

// 	t := reflect.TypeOf(a)
// 	generate(t)
// }

// func generate(t reflect.Type) {
// 	fmt.Println("Details: ", t.Name(), t.PkgPath(), t.Size(), t.String(), t.Kind())

// 	switch t.Kind() {
// 	case reflect.Uint8:
// 		genUint8(t)
// 	case reflect.Float64:
// 		genFloat64(t)
// 	case reflect.Struct:
// 		genStruct(t)
// 	default:
// 		fmt.Println("unhandled kind: ", t.Kind())
// 	}
// }

// func genStruct(t reflect.Type) {
// 	numField := t.NumField()
// 	fmt.Println("genStruct: ", numField)
// 	for i := 0; i < numField; i++ {
// 		field := t.Field(i)
// 		fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Offset, field.Index, field.Anonymous)
// 		generate(field.Type)
// 	}
// }

// func genFloat64(buf *bytes.Buffer, t reflect.Type) {
	
// }
// func genUint8(t reflect.Type) {
// 	fmt.Println("=====")
// }
