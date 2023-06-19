package alt

import (
	"fmt"
	"reflect"
	"bytes"
	"sort"
	"os"
	"io/fs"
	"go/format"

	// "github.com/unitoftime/flow/phy2"
	// "github.com/unitoftime/cod/alt/inner"
	// "github.com/unitoftime/cod/alt/test"
)

// TODO: technically a map from reflect.Kind to API names
// TODO: handle these: Complex64, Complex128
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
// func isStandard(name string) bool {
// 	if name == "" { return true } // Anything without a name isnt type defined
// 	_, std := supportedApis[name]
// 	return std
// }

func Generate(types ...any) {
	gen := newGenerator()
	gen.add(types...)
	buf := new(bytes.Buffer)
	gen.generate(buf)
	fmt.Println(buf.String())

	filename := "output.tmp"
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		os.WriteFile(filename, buf.Bytes(), fs.ModePerm)
		panic(err)
	}
	err = os.WriteFile(filename, formatted, fs.ModePerm)
	// err = os.WriteFile(filename, buf.Bytes(), fs.ModePerm)
	if err != nil {
		panic(err)
	}
}

type generator struct {
	Types map[reflect.Type]bool
}
func newGenerator() *generator {
	return &generator{
		Types: make(map[reflect.Type]bool),
	}
}

func (g *generator) add(types ...any) {
	for _, t := range types {
		g.Types[reflect.TypeOf(t)] = true
	}
}

func (g *generator) generate(buf *bytes.Buffer) {
	types := make([]reflect.Type, 0)
	for t := range g.Types {
		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Name() < types[j].Name()
	})

	fmt.Println("-----")
	fmt.Println(types)
	fmt.Println("-----")

	innerBuf := new(bytes.Buffer)
	for _, t := range types {
		// fmt.Println("Details: ", t.Name(), t.PkgPath(), t.String(), t.Kind())
		fmt.Println("Marsh/Unmarsh:", t.Name(), t.Kind())

		// Marshal
		innerBuf.Reset()
		g.innerMarshal(innerBuf, "t", t)

		err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
			"Type": t.Name(),
			"InnerCode": innerBuf.String(),
		})
		if err != nil { panic(err) }

		// Unmarshal
		innerBuf.Reset()
		g.innerUnmarshal(innerBuf, "t", t)

		err = BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
			"Type": t.Name(),
			"InnerCode": innerBuf.String(),
		})
		if err != nil { panic(err) }
	}
}

func (g *generator) innerMarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	switch t.Kind() {
	case reflect.Uint8: fallthrough
	case reflect.Uint16: fallthrough
	case reflect.Uint32: fallthrough
	case reflect.Uint64: fallthrough
	case reflect.Int8: fallthrough
	case reflect.Int16: fallthrough
	case reflect.Int32: fallthrough
	case reflect.Int64: fallthrough
	case reflect.Float32: fallthrough
	case reflect.Float64: fallthrough
	case reflect.Int: fallthrough
	case reflect.Uint: fallthrough
	case reflect.Bool: fallthrough
	case reflect.String:
		g.addBasicMarshal(buf, name, t)

	case reflect.Array:
		fmt.Println("array: ", name, t.Len())
		g.addArrayMarshal(buf, name, t)
		// g.innerMarshal(buf, name + "[i]", t.Elem())
	case reflect.Slice:
		fmt.Println("slice:", name)
		g.addSliceMarshal(buf, name, t)
	case reflect.Map:
		fmt.Println("map:", name)
		// g.innerMarshal(buf, name + "[k]", t.Key())
		// g.innerMarshal(buf, name + "v", t.Elem())
	case reflect.Pointer:
		fmt.Println("pointer:", name)
		// g.innerMarshal(buf, "*" + name,t.Elem())
	case reflect.Struct:
		fmt.Println("struct: ", name, t.NumField())
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			field := t.Field(i)
			// fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Type.Kind())
			g.innerMarshal(buf, name+"."+field.Name, field.Type)
		}
	default:
		fmt.Println("unhandled kind: ", name, t.Kind())
	}
}

func (g *generator) innerUnmarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	switch t.Kind() {
	case reflect.Uint8: fallthrough
	case reflect.Uint16: fallthrough
	case reflect.Uint32: fallthrough
	case reflect.Uint64: fallthrough
	case reflect.Int8: fallthrough
	case reflect.Int16: fallthrough
	case reflect.Int32: fallthrough
	case reflect.Int64: fallthrough
	case reflect.Float32: fallthrough
	case reflect.Float64: fallthrough
	case reflect.Int: fallthrough
	case reflect.Uint: fallthrough
	case reflect.Bool: fallthrough
	case reflect.String:
		g.addBasicUnmarshal(buf, name, t)

	case reflect.Array:
		fmt.Println("array: ", name, t.Len())
		// g.innerMarshal(buf, name + "[i]", t.Elem())
		g.addArrayUnmarshal(buf, name, t)
	case reflect.Slice:
		fmt.Println("slice:", name)
		// g.innerMarshal(buf, name + "[i]", t.Elem())
		g.addSliceUnmarshal(buf, name, t)
	case reflect.Map:
		fmt.Println("map:", name)
		// g.innerMarshal(buf, name + "[k]", t.Key())
		// g.innerMarshal(buf, name + "v", t.Elem())
	case reflect.Pointer:
		fmt.Println("pointer:", name)
		// g.innerMarshal(buf, "*" + name,t.Elem())
	case reflect.Struct:
		fmt.Println("struct: ", name, t.NumField())
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			field := t.Field(i)
			// fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Type.Kind())
			g.innerUnmarshal(buf, name+"."+field.Name, field.Type)
		}
	default:
		fmt.Println("unhandled kind: ", name, t.Kind())
	}
}

// --------------------------------------------------------------------------------
// Basic
// --------------------------------------------------------------------------------

func (g *generator) addBasicMarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)

	apiName, ok := supportedApis[t.Kind().String()]
	if !ok { panic(fmt.Sprintf("unknown kind: %s", t.Kind())) }

	err := BasicTemp.ExecuteTemplate(buf, "basic_marshal", map[string]any{
		"Name": name,
		"ApiName": apiName,
		"Type": t.Name(),
		"Cast": "", // TODO: Eventually (maybe not needed in this version)
	})
	if err != nil { panic(err) }
}


func (g *generator) addBasicUnmarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)

	apiName, ok := supportedApis[t.Kind().String()]
	if !ok { panic(fmt.Sprintf("unknown kind: %s", t.Kind())) }

	err := BasicTemp.ExecuteTemplate(buf, "basic_unmarshal", map[string]any{
		"Name": name,
		"ApiName": apiName,
		"Type": t.Name(),
		"Cast": "", // TODO: Eventually (maybe not needed in this version)
	})
	if err != nil { panic(err) }
}

// --------------------------------------------------------------------------------
// Array
// --------------------------------------------------------------------------------
func (g *generator) addArrayMarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)
	innerBuf := new(bytes.Buffer)
	idx := "i"
	innerName := fmt.Sprintf("%s[%s]", name, idx)
	g.innerMarshal(innerBuf, innerName, t.Elem())

	err := BasicTemp.ExecuteTemplate(buf, "array_marshal", map[string]any{
		"Index": idx,
		"Name": name,

		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}


func (g *generator) addArrayUnmarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)
	innerBuf := new(bytes.Buffer)
	idx := "i"
	innerName := fmt.Sprintf("%s[%s]", name, idx)
	g.innerUnmarshal(innerBuf, innerName, t.Elem())

	err := BasicTemp.ExecuteTemplate(buf, "array_unmarshal", map[string]any{
		"Index": idx,
		"Name": name,

		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

// --------------------------------------------------------------------------------
// Slice
// --------------------------------------------------------------------------------
func (g *generator) addSliceMarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)
	innerBuf := new(bytes.Buffer)
	idx := "i"
	innerName := fmt.Sprintf("%s[%s]", name, idx)
	g.innerMarshal(innerBuf, innerName, t.Elem())

	err := BasicTemp.ExecuteTemplate(buf, "slice_marshal", map[string]any{
		"Index": idx,
		"Name": name,

		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}


func (g *generator) addSliceUnmarshal(buf *bytes.Buffer, name string, t reflect.Type) {
	fmt.Println(t.Kind().String() + ":", name)
	innerBuf := new(bytes.Buffer)
	idx := "i"
	// innerName := fmt.Sprintf("%s[%s]", name, idx)
	innerName := "decoded"
	g.innerUnmarshal(innerBuf, innerName, t.Elem())

	err := BasicTemp.ExecuteTemplate(buf, "slice_unmarshal", map[string]any{
		"Index": idx,
		"Name": name,
		"Type": t.Name(),
		"VarName": innerName,
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

// // func (g *generator) addNewType(t reflect.Type) bool {
// // 	if isStandard(t.Name()) { return true }

// // 	_, exists := g.Types[t]
// // 	if exists { return false }
// // 	g.Types[t] = true
// // 	return true
// // }

// func (g *generator) search(t reflect.Type) {

// 	success := g.addNewType(t)
// 	if !success {
// 		fmt.Println("Type already exists. skipping")
// 	}
// 	fmt.Println("Details: ", t.Name(), t.PkgPath(), t.String(), t.Kind())


// 	switch t.Kind() {
// 	case reflect.Struct:
// 		numField := t.NumField()
// 		fmt.Println("genStruct: ", numField)
// 		for i := 0; i < numField; i++ {
// 			field := t.Field(i)
// 			fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Type.Kind())
// 			g.search(field.Type)
// 		}
// 	default:
// 		fmt.Println("unhandled kind: ", t.Kind())
// 	}
// }

// func (g *generator) generate(buf *bytes.Buffer) {
// 	for t := range g.Types {
// 		g.marshalFunc(buf, t)
// 		g.unmarshalFunc(buf, t)
// 	}
// }


// func (g *generator) marshalFunc(buf *bytes.Buffer, t reflect.Type) {
// 	// inner := new(bytes.Buffer)

// 	if t.Kind() == reflect.Struct {
// 		numField := t.NumField()
// 		for i := 0; i < numField; i++ {
// 			field := t.Field(i)
// 			g.marshalField(buf, field)
// 		}
// 	}

// 	// err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
// 	// 	"Type": t.Name(),
// 	// 	"InnerCode": inner.String(),
// 	// })
// 	// if err != nil { panic(err) }
// }

// func (g *generator) unmarshalFunc(buf *bytes.Buffer, t reflect.Type) {
// 	// err := BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
// 	// 	"Type": t.Name(),
// 	// 	"InnerCode": "TODO",
// 	// })
// 	// if err != nil { panic(err) }
// }

// func (g *generator) marshalField(buf *bytes.Buffer, field reflect.StructField) {
// 	fmt.Println("marshalField: ", field.Name, field.Type.Kind())

// 	if field.Type.Kind() == reflect.Struct {
// 		g.marshalFunc(buf, field.Type)
// 	}

// }

// // func (g *generator) unmarshalFields(buf *bytes.Buffer, t reflect.Type) {
// // 	err := BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
// // 		"Type": t.Name(),
// // 		"InnerCode": "TODO",
// // 	})
// // 	if err != nil { panic(err) }
// // }

// // func generateAny(a any) {

// // 	fmt.Println("")
// // 	fmt.Printf("generate: %T\n", a)

// // 	t := reflect.TypeOf(a)
// // 	generate(t)
// // }

// // func generate(t reflect.Type) {
// // 	fmt.Println("Details: ", t.Name(), t.PkgPath(), t.Size(), t.String(), t.Kind())

// // 	switch t.Kind() {
// // 	case reflect.Uint8:
// // 		genUint8(t)
// // 	case reflect.Float64:
// // 		genFloat64(t)
// // 	case reflect.Struct:
// // 		genStruct(t)
// // 	default:
// // 		fmt.Println("unhandled kind: ", t.Kind())
// // 	}
// // }

// // func genStruct(t reflect.Type) {
// // 	numField := t.NumField()
// // 	fmt.Println("genStruct: ", numField)
// // 	for i := 0; i < numField; i++ {
// // 		field := t.Field(i)
// // 		fmt.Println("Field: ", field.Name, field.PkgPath, field.Type, field.Tag, field.Offset, field.Index, field.Anonymous)
// // 		generate(field.Type)
// // 	}
// // }

// // func genFloat64(buf *bytes.Buffer, t reflect.Type) {
	
// // }
// // func genUint8(t reflect.Type) {
// // 	fmt.Println("=====")
// // }
