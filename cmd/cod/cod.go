package main

import (
	"fmt"
	"bytes"
	"os"
	"strings"
	// "math"
	"io/fs"
	"go/ast"
	"go/parser"
	// "go/printer"
	"go/token"
	"go/format"
	"text/template"
	"strconv"
	"path/filepath"


	// _ "embed"
)

// List of supported reads and writes
var supportedApis = map[string]string{
	"uint8": "Uint8",
	"int8": "Int8",

	// As a default, we always use Variable length encoding APIs for anything > 2 bytes
	"uint16": "VarUint16",
	"uint32": "VarUint32",
	"uint64": "VarUint64",

	"int16": "VarInt16",
	"int32": "VarInt32",
	"int64": "VarInt64",

	"string": "String",
	"bool": "Bool",
}


func main() {
	generatePackage(".")
}

func generatePackage(dir string) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range packages {
		fmt.Println("Parsing", pkg.Name)
		bv := &Visitor{
			// buf: &bytes.Buffer{},
			pkg: pkg,
			fset: fset,
			// lastCommentPos: &tokenStart,
			structs: make(map[string]StructData),
			imports: make(map[string]string),
			usedImports: make(map[string]bool),
		}

		// We start walking our Visitor `bv` through the AST in a depth-first way.
		ast.Walk(bv, pkg)

		bv.Output("cod_encode.go")
	}
}
func (v *Visitor) formatGen(decl ast.GenDecl, cGroups []*ast.CommentGroup) (StructData, bool) {
	structData := StructData{}

	// Skip everything that isn't a type. we can only generate for types
	if decl.Tok != token.TYPE {
		decl.Doc = nil // nil the Doc field so that we don't print it
		return structData, false
	}

	fields := make([]Field, 0)

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			directive, directiveCSV := getDirective(decl)
			if directive == DirectiveNone {
				return structData, false
			}

			// fmt.Println("TypeSpec: ", s.Name.Name)
			structData.Name = s.Name.Name
			structData.Directive = directive
			structData.DirectiveCSV = directiveCSV

			// fmt.Printf("Struct Type: %T\n", s.Type)
			sType, ok := s.Type.(*ast.StructType)
			if !ok {
				// Not a struct, then its an alias. So handle that if we can
				name := "t"
				idxDepth := 0
				field := v.generateField(name, idxDepth+1, s.Type)
				fields = append(fields, &AliasField{
					Name: name,
					AliasType: s.Name.Name,
					Field: field,
					IndexDepth: idxDepth,
				})
				continue
			}

			for _, f := range sType.Fields.List {
				for _, n := range f.Names {
					// fmt.Println("Field: ", n.Name, f.Type, f.Tag)
					// fmt.Printf("%T\n", f.Type)

					field := v.generateField("t." + n.Name, 0, f.Type)
					if f.Tag != nil {
						field.SetTag(f.Tag.Value)
					}

					fields = append(fields, field)
				}
			}
		}
	}
	structData.Fields = fields

	return structData, true
}

func (v *Visitor) generateField(name string, idxDepth int, node ast.Node) Field {
	switch expr := node.(type) {
	case *ast.Ident:
		// fmt.Println("Ident: ", expr.Name)
		field := &BasicField{
			Name: name,
			Type: expr.Name,
		}

		return field

	// case *ast.StarExpr:
	// 	// fmt.Println("StarExpr: ", expr.Name)
	// 	fmt.Printf("STAR %T\n", expr.X)
	// 	field := generateField(name, 0, expr.X) // TODO: idxDepth???
	// 	return PointerField{
	// 		Field: field,
	// 	}

	case *ast.ArrayType:
		// fmt.Printf("ARRAY %T %T\n", expr.Len, expr.Elt)

		if expr.Len == nil {
			idxString := fmt.Sprintf("[i%d]", idxDepth)
			field := v.generateField(name + idxString, idxDepth + 1, expr.Elt)
			return &SliceField{
				Name: name,
				// Type: field.GetType(),
				Field: field,
				IndexDepth: idxDepth,
			}
		} else {
			idxString := fmt.Sprintf("[i%d]", idxDepth)
			field := v.generateField(name + idxString, idxDepth + 1, expr.Elt)

			lString := expr.Len.(*ast.BasicLit).Value
			length, err := strconv.Atoi(lString)
			if err != nil { panic("ERR") }
			return &ArrayField{
				Name: name,
				Len: length,
				Field: field,
				IndexDepth: idxDepth,
			}
		}

	case *ast.MapType:
		// fmt.Printf("MAP %T %T\n", expr.Key, expr.Value)
		keyString := fmt.Sprintf("[k%d]", idxDepth)
		valString := fmt.Sprintf("[v%d]", idxDepth)
		key := v.generateField(name + keyString, idxDepth + 1, expr.Key)
		val := v.generateField(name + valString, idxDepth + 1, expr.Value)
		return &MapField{
			Name: name,
			Key: key,
			Val: val,
			IndexDepth: idxDepth,
		}
	case *ast.SelectorExpr:
		// Note: anything that is a selector expression (ie phy.Position) is guaranteed to be a struct. so it must implement the required struct interface
		fmt.Printf("SELECTOREXPR: %T %T\n", expr.X, expr.Sel)
		x := expr.X.(*ast.Ident)
		fmt.Println("SELECTOREXPR:", x.Name, expr.Sel.Name)
		field := &BasicField{
			Name: name,
			// Type: "UNKNOWN_SELECTOR_EXPR", // This will force it to resolve to the struct marshaller
			Type: x.Name + "." + expr.Sel.Name, // This will force it to resolve to the struct marshaller
		}

		v.usedImports[x.Name] = true // Store the import name so we can pull it later

		return field

	default:
		panic(fmt.Sprintf("unknown type %T", expr))
	}

	return nil
}

func getDirective(t ast.GenDecl) (DirectiveType, []string) {
	if t.Doc == nil {
		return DirectiveNone, nil
	}

	for _, c := range t.Doc.List {
		after, foundStruct := strings.CutPrefix(c.Text, "//cod:struct")
		if foundStruct {
			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			return DirectiveStruct, csv
		}

		after, foundUnion := strings.CutPrefix(c.Text, "//cod:union")
		if foundUnion {
			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			return DirectiveUnion, csv
		}
	}
	return DirectiveNone, nil
}

type Visitor struct {
	pkg *ast.Package  // The package that we are processing
	fset *token.FileSet // The fileset of the package we are processing
	file *ast.File // The file we are currently processing (Can be nil if we haven't started processing a file yet!)
	cmap ast.CommentMap // The comment map of the file we are processing

	structs map[string]StructData
	imports map[string]string // Maps a selector source to a package path
	usedImports map[string]bool // List of encoded selector expressions
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil { return nil }

	// fmt.Printf("Node: %T\n", node)

	// If we are a package, then just keep searching
	_, ok := node.(*ast.Package)
	if ok { return v }

	// If we are a file, then store some data in the visitor so we can use it later
	file, ok := node.(*ast.File)
	if ok {
		v.file = file
		v.cmap = ast.NewCommentMap(v.fset, file, file.Comments)

		for _, importSpec := range file.Imports {
			fmt.Println("IMPORT SPEC: ", importSpec)
			path := importSpec.Path.Value

			name := strings.TrimSuffix(filepath.Base(path), `"`)

			// If there was a custom name, use that
			nameIdent := importSpec.Name
			if nameIdent != nil {
				name = nameIdent.Name
			}
			fmt.Println("IMPORT: ", name, path)
			v.imports[name] = path
		}
		return v
	}

	// If we are a function, do the function formatting
	_, ok = node.(*ast.FuncDecl)
	if ok {
		return nil // Skip: we don't handle funcs
	}

	gen, ok := node.(*ast.GenDecl)
	if ok {
		cgroups := v.cmap.Filter(gen).Comments()
		sd, ok := v.formatGen(*gen, cgroups)
		if ok {
			v.structs[sd.Name] = sd
		}


		return nil
	}

	// If all else fails, then keep looking
	return v
}

type StructData struct {
	Name string
	Directive DirectiveType
	DirectiveCSV []string
	Fields []Field
}

func (s *StructData) WriteStructMarshal(buf *bytes.Buffer) {
	if s.Directive == DirectiveStruct {
		for _, f := range s.Fields {
			f.WriteMarshal(buf)
		}
	} else if s.Directive == DirectiveUnion {
		innerBuf := new(bytes.Buffer)
		for tag, name := range s.DirectiveCSV {
			err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_marshal", map[string]any{
				"Type": name,
				"Tag": tag+1,
				// "InnerCode": string(innerInnerBuf.Bytes()),
				// For now I'm just going to use the requirement that you can only add items to the union that implement the EncodeCod function pair. But you could fix this and let in primitives. It just gets hard and isnt as useful for me right now
				// "InnerCode": "t.EncodeCod(buf)",
			})
			if err != nil { panic(err) }
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_marshal", map[string]any{
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }

	}
}

func (s *StructData) WriteStructUnmarshal(buf *bytes.Buffer) {
	if s.Directive == DirectiveStruct {
		for _, f := range s.Fields {
			f.WriteUnmarshal(buf)
		}
	} else if s.Directive == DirectiveUnion {
		innerBuf := new(bytes.Buffer)
		for tag, name := range s.DirectiveCSV {
			err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_unmarshal", map[string]any{
				"Type": name,
				"Tag": tag+1,
				// "InnerCode": "t.EncodeCod(buf)",
			})
			if err != nil { panic(err) }
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_unmarshal", map[string]any{
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}
}

type Field interface {
	WriteMarshal(*bytes.Buffer)
	WriteUnmarshal(*bytes.Buffer)
	SetTag(string)
	SetName(string)
	GetType() string
}

type BasicField struct {
	Name string
	Type string
	Tag string
}

func (f *BasicField) SetName(name string) {
	f.Name = name
}

func (f *BasicField) SetTag(tag string) {
	f.Tag = tag
}
func (f *BasicField) GetType() string {
	return f.Type
}

func (f BasicField) WriteMarshal(buf *bytes.Buffer) {
	apiName, supported := supportedApis[f.Type]
	if supported {
		err := BasicTemp.ExecuteTemplate(buf, "basic_marshal", map[string]any{
			"Name": f.Name,
			"ApiName": apiName,
		})
		if err != nil { panic(err) }
	} else {
		// fmt.Println("Found Struct: ", f.Name)
		err := BasicTemp.ExecuteTemplate(buf, "struct_marshal", map[string]any{
			"Name": f.Name,
		})
		if err != nil { panic(err) }
	}
}

func (f BasicField) WriteUnmarshal(buf *bytes.Buffer) {
	apiName, supported := supportedApis[f.Type]
	if supported {
		err := BasicTemp.ExecuteTemplate(buf, "basic_unmarshal", map[string]any{
			"Name": f.Name,
			"ApiName": apiName,
		})
		if err != nil { panic(err) }
	} else {
		// fmt.Println("Found Struct: ", f.Name)
		err := BasicTemp.ExecuteTemplate(buf, "struct_unmarshal", map[string]any{
			"Name": f.Name,
			"Type": f.GetType(),
		})
		if err != nil { panic(err) }
	}


	// pointerStar := "reg"
	// if f.Pointer { pointerStar = "ptr" }

	// templateName := fmt.Sprintf("%s_%s_unmarshal", pointerStar, f.Type)
	// err := BasicTemp.ExecuteTemplate(buf, templateName, map[string]any{
	// 	"Name": f.Name,
	// })
	// if err != nil {
	// 	fmt.Println("Couldn't find type, assuming its a struct: ", f.Name)
	// 	templateName := fmt.Sprintf("%s_%s_unmarshal", pointerStar, "struct")
	// 	err := BasicTemp.ExecuteTemplate(buf, templateName, map[string]any{
	// 		"Name": f.Name,
	// 	})
	// 	if err != nil { panic(err) }
	// }
}

type ArrayField struct {
	Name string
	Field Field
	Len int
	Tag string
	IndexDepth int
}

func (f *ArrayField) SetName(name string) {
	f.Name = name
}
func (f *ArrayField) SetTag(tag string) {
	f.Tag = tag
}
func (f *ArrayField) GetType() string {
	return fmt.Sprintf("[%d]%s", f.Len, f.Field.GetType())
}

func (f ArrayField) WriteMarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	f.Field.WriteMarshal(innerBuf)


	err := BasicTemp.ExecuteTemplate(buf, "array_marshal", map[string]any{
		"Name": f.Name,
		"Index": fmt.Sprintf("i%d", f.IndexDepth),
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}


func (f ArrayField) WriteUnmarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	f.Field.WriteUnmarshal(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "array_unmarshal", map[string]any{
		"Name": f.Name,
		"Index": fmt.Sprintf("i%d", f.IndexDepth),
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }

}

type SliceField struct {
	Name string
	// Type string
	Field Field
	Tag string
	IndexDepth int
}

func (f *SliceField) SetName(name string) {
	f.Name = name
}
func (f *SliceField) SetTag(tag string) {
	f.Tag = tag
}
func (f *SliceField) GetType() string {
	return fmt.Sprintf("[]%s", f.Field.GetType())
}

func (f SliceField) WriteMarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	idxVar := fmt.Sprintf("i%d", f.IndexDepth)
	f.Field.SetName(fmt.Sprintf("%s[%s]", f.Name, idxVar))
	f.Field.WriteMarshal(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "slice_marshal", map[string]any{
		"Name": f.Name,
		"Type": f.Field.GetType(),
		"Index": idxVar,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}


func (f SliceField) WriteUnmarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	varName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(varName)
	f.Field.WriteUnmarshal(innerBuf)

	// fmt.Println("GETTYPE: ", f.Field.GetType())
	err := BasicTemp.ExecuteTemplate(buf, "slice_unmarshal", map[string]any{
		"Name": f.Name,
		"VarName": varName,
		"Type": f.Field.GetType(),
		"Index": fmt.Sprintf("i%d", f.IndexDepth),
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }

}

type MapField struct {
	Name string
	Key Field
	Val Field
	Tag string
	IndexDepth int
}

func (f *MapField) SetName(name string) {
	f.Name = name
}
func (f *MapField) SetTag(tag string) {
	f.Tag = tag
}
func (f *MapField) GetType() string {
	return fmt.Sprintf("map[%s]%s", f.Key.GetType(), f.Val.GetType())
}

func (f MapField) WriteMarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	keyIdxName := fmt.Sprintf("k%d", f.IndexDepth)
	f.Key.SetName(keyIdxName)
	f.Key.WriteMarshal(innerBuf)

	valIdxName := fmt.Sprintf("v%d", f.IndexDepth)
	f.Val.SetName(valIdxName)
	f.Val.WriteMarshal(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "map_marshal", map[string]any{
		"Name": f.Name,
		"KeyIdx": keyIdxName,
		"ValIdx": valIdxName,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}


func (f MapField) WriteUnmarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	keyVarName := fmt.Sprintf("key%d", f.IndexDepth)
	f.Key.SetName(keyVarName)
	f.Key.WriteUnmarshal(innerBuf)

	valVarName := fmt.Sprintf("val%d", f.IndexDepth)
	f.Val.SetName(valVarName)
	f.Val.WriteUnmarshal(innerBuf)

	// fmt.Println("GETTYPE: ", f.GetType(), f.Key.GetType(), f.Val.GetType())
	err := BasicTemp.ExecuteTemplate(buf, "map_unmarshal", map[string]any{
		"Name": f.Name,
		"Type": f.GetType(),
		"KeyVar": keyVarName,
		"ValVar": valVarName,
		"KeyType": f.Key.GetType(),
		"ValType": f.Val.GetType(),
		"InnerCode": string(innerBuf.Bytes()),

		"Index": fmt.Sprintf("i%d", f.IndexDepth),
	})
	if err != nil { panic(err) }

}

type AliasField struct {
	Name string
	AliasType string
	Field Field
	Tag string
	IndexDepth int
}

func (f *AliasField) SetName(name string) {
	f.Name = name
}
func (f *AliasField) SetTag(tag string) {
	f.Tag = tag
}
func (f *AliasField) GetType() string {
	return fmt.Sprintf("%s", f.Field.GetType())
}

func (f AliasField) WriteMarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteMarshal(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "alias_marshal", map[string]any{
		"Name": f.Name,
		"AliasType": f.Name,
		"Type": f.GetType(),
		"ValName": valName,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}


func (f AliasField) WriteUnmarshal(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteUnmarshal(innerBuf)

	// fmt.Println("ALIAS_GETTYPE: ", f.GetType(), f.Field.GetType())
	err := BasicTemp.ExecuteTemplate(buf, "alias_unmarshal", map[string]any{
		"Name": f.Name,
		"AliasType": f.AliasType,
		"Type": f.GetType(),
		"ValName": valName,
		"ValType": f.Field.GetType(),
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }

}

// type FieldType uint16
// const (
// 	FieldNone FieldType = iota
// 	FieldUint8
// 	FieldUint16
// 	FieldUint32
// 	FieldUint64
// 	FieldInt8
// 	FieldInt16
// 	FieldInt32
// 	FieldInt64

// 	FieldString
// 	FieldStruct
// )
type DirectiveType uint8
const (
	DirectiveNone DirectiveType = iota
	DirectiveStruct
	DirectiveUnion
)


func (v *Visitor) Output(filename string) {
	marshal, err := template.New("marshal").Parse(`
func (t {{.Name}})EncodeCod(bs []byte) []byte {
{{.MarshalCode}}
return bs
}
`)
	if err != nil { panic(err) }

	unmarshal, err := template.New("unmarshal").Parse(`
func (t *{{.Name}})DecodeCod(bs []byte) (int, error) {
var err error
var n int
var nOff int

{{.MarshalCode}}

return n, err
}
`)
	if err != nil { panic(err) }

	importCod := false
	for _, sd := range v.structs {
		if sd.Directive == DirectiveUnion {
			importCod = true
			break
		}
	}

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("package " + v.pkg.Name)
		buf.WriteString(`
import (
	"github.com/unitoftime/cod/backend"
`)

	if importCod {
		buf.WriteString(`
	"github.com/unitoftime/cod"
`)
	}

	for k := range v.usedImports {
		if k == "cod" { continue }
		fmt.Println("Used Import: ", k)
		path, ok := v.imports[k]
		fmt.Println("Used Import: ", k, path, ok)
		if !ok {
			panic("couldnt find import!")
		}
		buf.WriteString("\n"+path+"\n")
	}
		buf.WriteString(`
)`)

	marshBuf := bytes.NewBuffer([]byte{})
	unmarshBuf := bytes.NewBuffer([]byte{})
	for _, sd := range v.structs {
		marshBuf.Reset()
		unmarshBuf.Reset()

		fmt.Println("Struct: ", sd.Name)

		// Write the marshal code
		sd.WriteStructMarshal(marshBuf)

		// Write the unmarshal code
		sd.WriteStructUnmarshal(unmarshBuf)

		// Write the encode func
		err = marshal.Execute(buf, map[string]any{
			"Name": sd.Name,
			"MarshalCode": string(marshBuf.Bytes()),
		})
		if err != nil { panic(err) }

		// Write the decode func
		err = unmarshal.Execute(buf, map[string]any{
			"Name": sd.Name,
			"MarshalCode": string(unmarshBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}

	for _, sd := range v.structs {
		if sd.Directive != DirectiveUnion { continue }

		// Create constructors, getters, setters per union type
		err := BasicTemp.ExecuteTemplate(buf, "union_getter", map[string]any{
			"Name": sd.Name,
		})
		if err != nil { panic(err) }
		err = BasicTemp.ExecuteTemplate(buf, "union_setter", map[string]any{
			"Name": sd.Name,
		})
		if err != nil { panic(err) }
		err = BasicTemp.ExecuteTemplate(buf, "union_constructor", map[string]any{
			"Name": sd.Name,
		})
		if err != nil { panic(err) }
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		err = os.WriteFile(filename, buf.Bytes(), fs.ModePerm)
		if err != nil {
			panic(err)
		}
		return
	}

	err = os.WriteFile(filename, formatted, fs.ModePerm)
	// err = os.WriteFile(filename, buf.Bytes(), fs.ModePerm)
	if err != nil {
		panic(err)
	}
}


