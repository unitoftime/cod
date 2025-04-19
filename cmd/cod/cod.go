package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"

	"go/ast"
	"go/parser"

	"go/token"

	"path/filepath"
)

// List of supported reads and writes
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

func main() {
	now := time.Now()
	defer printDuration("main", now)

	// newMain()
	// generatePackage(".")
	generateAll(".")
}

func generateAll(dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() { return nil } // Skip if not the directory

		fullPath := filepath.Join(dir, path)
		fmt.Println("generatePackage:", fullPath)
		generatePackage(fullPath)
		return nil
	})
}

func generatePackage(dir string) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range packages {
		fmt.Println("Parsing Package:", pkg.Name)
		bv := &Visitor{
			// buf: &bytes.Buffer{},
			pkg: pkg,
			fset: fset,
			// lastCommentPos: &tokenStart,
			requests: make(map[string][]GenRequest),

			structs: make(map[string]StructData),
			imports: make(map[string]string),
			usedImports: make(map[string]bool),
		}

		// We start walking our Visitor `bv` through the AST in a depth-first way.
		ast.Walk(bv, pkg)

		bv.Output(filepath.Join(dir, "cod_encode.go"))
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
			add := v.getRequests(s.Name.Name, decl)

			directive, directiveCSV := v.getDirective(decl)
			if directive == DirectiveNone {
				if add {
					structData.Name = s.Name.Name
					return structData, true
				}

				return structData, false
			}

			debugPrintln("TypeSpec: ", s.Name.Name)
			debugPrintf("TypeSpec: %T\n", s.Type)
			structData.Name = s.Name.Name
			structData.Directive = directive
			structData.DirectiveCSV = directiveCSV

			// debugPrintf("Struct Type: %T\n", s.Type)
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

			debugPrintln("Fields: ", sType.Fields.List)

			unionTag := 1
			for _, f := range sType.Fields.List {
				if f.Names == nil {
					debugPrintln("UnnamedField: ", f.Type, f.Tag)

					// TODO: probably come up with a better way of getting the name
					name := ""
					fIdent, ok := f.Type.(*ast.Ident)
					if ok {
						name = fIdent.Name
					}
					sel, ok := f.Type.(*ast.SelectorExpr)
					if ok {
						x := sel.X.(*ast.Ident)
						name = x.Name + "." + sel.Sel.Name
					}

					idxDepth := 0
					field := v.generateField("t." + name, idxDepth+1, f.Type)
					if f.Tag != nil {
						field.SetTag(f.Tag.Value)
					}

					if directive == DirectiveUnionDef {
						field = &UnionField{
							Name: name,
							UnionTag: unionTag,
							Field: field,
							IndexDepth: idxDepth,
						}

						unionTag++
					}

					fields = append(fields, field)
				} else {
					for _, n := range f.Names {
						debugPrintln("Field: ", n.Name, f.Type, f.Tag)
						debugPrintf("%T\n", f.Type)

						idxDepth := 0
						field := v.generateField("t." + n.Name, idxDepth+1, f.Type)
						if f.Tag != nil {
							field.SetTag(f.Tag.Value)
						}

						if directive == DirectiveUnionDef {
							field = &UnionField{
								Name: n.Name,
								UnionTag: unionTag,
								Field: field,
								IndexDepth: idxDepth,
							}

							unionTag++
						}

						fields = append(fields, field)
					}
				}
			}
		}
	}
	structData.Fields = fields

	return structData, true
}

func (v *Visitor) generateField(name string, idxDepth int, node ast.Node) Field {
	debugPrintf("generateField: %T\n", node)

	switch expr := node.(type) {
	case *ast.Ident:
		debugPrintln("Ident: ", expr.Name)
		field := &BasicField{
			Name: name,
			Type: expr.Name,
		}

		return field

	case *ast.StarExpr:
		// debugPrintln("StarExpr: ", expr.Name)
		debugPrintf("STAR %T\n", expr.X)
		field := v.generateField(name, idxDepth+1, expr.X)
		return &PointerField{
			Name: name,
			Field: field,
			IndexDepth: idxDepth,
		}

	case *ast.ArrayType:
		// debugPrintf("ARRAY %T %T\n", expr.Len, expr.Elt)

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

			lengthIdentName := ""

			lString, ok := expr.Len.(*ast.Ident)
			if ok {
				lengthIdentName = lString.Name
			} else {
				blString, ok := expr.Len.(*ast.BasicLit)
				if ok {
					lengthIdentName = blString.Value
				}
			}

			if lengthIdentName == "" {
				panic("could not find array identifier")
			}

			return &ArrayField{
				Name: name,
				Len: lengthIdentName,
				Field: field,
				IndexDepth: idxDepth,
			}
		}

	case *ast.MapType:
		// debugPrintf("MAP %T %T\n", expr.Key, expr.Value)
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
		debugPrintf("SELECTOREXPR: %T %T\n", expr.X, expr.Sel)
		x := expr.X.(*ast.Ident)
		debugPrintln("SELECTOREXPR:", x.Name, expr.Sel.Name)
		field := &BasicField{
			Name: name,
			// Type: "UNKNOWN_SELECTOR_EXPR", // This will force it to resolve to the struct marshaller
			Type: x.Name + "." + expr.Sel.Name, // This will force it to resolve to the struct marshaller
		}

		v.usedImports[x.Name] = true // Store the import name so we can pull it later

		return field

	default:
		panic(fmt.Sprintf("%s:, unknown type %T", name, expr))
	}

	return nil
}

func (v *Visitor) getRequests(name string, t ast.GenDecl) bool {
	if t.Doc == nil {
		return false
	}

	for _, c := range t.Doc.List {
		after, found := strings.CutPrefix(c.Text, "//cod:component")
		if found {
			v.usedImports["ecs"] = true
			v.imports["ecs"] = "\"github.com/unitoftime/ecs\""

			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			v.requests[name] = append(v.requests[name], GenRequest{
				Type: RequestTypeComponent,
				CSV: csv,
			})
		}
	}
	return (len(v.requests[name]) > 0)
}

func (v *Visitor) getDirective(t ast.GenDecl) (DirectiveType, []string) {
	if t.Doc == nil {
		return DirectiveNone, nil
	}

	for _, c := range t.Doc.List {
		after, foundStruct := strings.CutPrefix(c.Text, "//cod:struct")
		if foundStruct {
			v.usedImports["backend"] = true
			v.imports["backend"] = "\"github.com/unitoftime/cod/backend\""

			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			return DirectiveStruct, csv
		}

		after, foundUnion := strings.CutPrefix(c.Text, "//cod:union")
		if foundUnion {
			v.usedImports["fmt"] = true // Some panic statements require fmt
			v.imports["fmt"] = "\"fmt\""

			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			return DirectiveUnion, csv
		}

		after, foundUnionDef := strings.CutPrefix(c.Text, "//cod:def")
		if foundUnionDef {
			csv := strings.Split(after, ",")
			for i := range csv {
				csv[i] = strings.TrimSpace(csv[i])
			}

			return DirectiveUnionDef, csv
		}
	}
	return DirectiveNone, nil
}

type RequestType int
const (
	RequestTypeComponent RequestType = iota
)
type GenRequest struct {
	Type RequestType
	CSV []string
}
type StructData2 struct {
	Name string
	Fields []Field
}

type Visitor struct {
	pkg *ast.Package  // The package that we are processing
	fset *token.FileSet // The fileset of the package we are processing
	file *ast.File // The file we are currently processing (Can be nil if we haven't started processing a file yet!)
	cmap ast.CommentMap // The comment map of the file we are processing

	requests map[string][]GenRequest
	structs2 map[string]StructData2

	structs map[string]StructData

	imports map[string]string // Maps a selector source to a package path
	usedImports map[string]bool // List of encoded selector expressions
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil { return nil }

	// debugPrintf("Node: %T\n", node)

	// If we are a package, then just keep searching
	_, ok := node.(*ast.Package)
	if ok { return v }

	// If we are a file, then store some data in the visitor so we can use it later
	file, ok := node.(*ast.File)
	if ok {
		v.file = file
		v.cmap = ast.NewCommentMap(v.fset, file, file.Comments)

		for _, importSpec := range file.Imports {
			debugPrintln("IMPORT SPEC: ", importSpec)
			path := importSpec.Path.Value

			name := strings.TrimSuffix(filepath.Base(path), `"`)

			// If there was a custom name, use that
			nameIdent := importSpec.Name
			if nameIdent != nil {
				name = nameIdent.Name
			}
			debugPrintln("IMPORT: ", name, path)
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

func (v *Visitor) WriteUnionCodeToBuffer(s *StructData, buf io.Writer) {
	if s.Directive != DirectiveUnion { return } // Exit if not union

	// For unions we lookup the union def which must be the first csv element
	unionDefName := s.DirectiveCSV[0]
	unionDef, ok := v.structs[unionDefName]
	if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

	debugPrintln("UDEF: ", unionDef.Fields)
	// GetTag()
	{
		innerBuf := new(bytes.Buffer)
		for _, fInterface := range unionDef.Fields {
			f := fInterface.(*UnionField)
			err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_get_tag", map[string]any{
				"Name": f.Name,
				"Type": f.GetType(),
				"Tag": f.UnionTag,
			})
			if err != nil { panic(err) }
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_get_tag_func", map[string]any{
			"Name": s.Name,
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}

	// GetSize()
	{
		err := BasicTemp.ExecuteTemplate(buf, "union_get_size_func", map[string]any{
			"Name": s.Name,
			"Size": len(unionDef.Fields) + 1, // Note: + 1 b/c 0 is the nil case
		})
		if err != nil { panic(err) }
	}
}


func (v *Visitor) WriteEqualityCodeToBuffer(s *StructData, buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	if s.Directive == DirectiveStruct {
		for _, f := range s.Fields {
			f.WriteEquality(innerBuf)
		}
		// Write the equality func
		err := BasicTemp.ExecuteTemplate(buf, "equality_func", map[string]any{
			"Name": s.Name,
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }

	} else if s.Directive == DirectiveUnion {
		// For unions we lookup the union def which must be the first csv element
		unionDefName := s.DirectiveCSV[0]
		unionDef, ok := v.structs[unionDefName]
		if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

		for _, fInterface := range unionDef.Fields {
			f := fInterface.(*UnionField)
			err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_equality", map[string]any{
				"Name": f.Name,
				"Name2": "t"+f.Name,
				"Type": f.GetType(),
				"Tag": f.UnionTag,
			})
			if err != nil { panic(err) }
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_equality_func", map[string]any{
			"Name": s.Name,
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}
}


func (v *Visitor) WriteStructMarshal(s *StructData, buf *bytes.Buffer) {
	if s.Directive == DirectiveStruct {
		for _, f := range s.Fields {
			f.WriteMarshal(buf)
		}
	} else if s.Directive == DirectiveUnion {
		// For unions we lookup the union def which must be the first csv element
		unionDefName := s.DirectiveCSV[0]
		unionDef, ok := v.structs[unionDefName]
		if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

		innerBuf := new(bytes.Buffer)
		debugPrintln("UDEF: ", unionDef.Fields)
		for _, f := range unionDef.Fields {
			f.WriteMarshal(innerBuf)
		}
		err := BasicTemp.ExecuteTemplate(buf, "union_marshal", map[string]any{
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}
}

func (v *Visitor) WriteStructUnmarshal(s *StructData, buf *bytes.Buffer) {
	if s.Directive == DirectiveStruct {
		for _, f := range s.Fields {
			f.WriteUnmarshal(buf)
		}
	} else if s.Directive == DirectiveUnion {
		// For unions we lookup the union def which must be the first csv element
		unionDefName := s.DirectiveCSV[0]
		unionDef, ok := v.structs[unionDefName]
		if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

		innerBuf := new(bytes.Buffer)
		for _, f := range unionDef.Fields {
			f.WriteUnmarshal(innerBuf)
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_unmarshal", map[string]any{
			"InnerCode": string(innerBuf.Bytes()),
		})
		if err != nil { panic(err) }
	}
}

type Field interface {
	WriteEquality(*bytes.Buffer)
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

func (f BasicField) WriteEquality(buf *bytes.Buffer) {
	skip := tagSearchSkip(f.Tag)
	debugPrintln("Skip: ", skip)
	cast := tagSearchCast(f.Tag)
	debugPrintln("Cast: ", cast)

	// Don't add if this is set to skip
	if shouldSkipEquality(skip) {
		return
	}


	apiType := f.Type
	if cast != "" {
		apiType = cast
	}

	apiName, supported := supportedApis[apiType]
	if supported {
		err := BasicTemp.ExecuteTemplate(buf, "basic_equality", map[string]any{
			"Name": f.Name,
			"Name2": "t"+f.Name,
			"ApiName": apiName,
		})
		if err != nil { panic(err) }
	} else {
		// debugPrintln("Found Struct: ", f.Name)
		err := BasicTemp.ExecuteTemplate(buf, "struct_equality", map[string]any{
			"Name": f.Name,
			"Name2": "t"+f.Name,
		})
		if err != nil { panic(err) }
	}
}

func (f BasicField) WriteMarshal(buf *bytes.Buffer) {
	skip := tagSearchSkip(f.Tag)
	debugPrintln("Skip: ", skip)
	cast := tagSearchCast(f.Tag)
	debugPrintln("Cast: ", cast)

	// Don't add if this is set to skip
	if shouldSkipSerdes(skip) {
		return
	}

	apiType := f.Type
	if cast != "" {
		apiType = cast
	}

	apiName, supported := supportedApis[apiType]
	if supported {
		err := BasicTemp.ExecuteTemplate(buf, "basic_marshal", map[string]any{
			"Name": f.Name,
			"ApiName": apiName,
			"Cast": cast,
		})
		if err != nil { panic(err) }
	} else {
		// debugPrintln("Found Struct: ", f.Name)
		err := BasicTemp.ExecuteTemplate(buf, "struct_marshal", map[string]any{
			"Name": f.Name,
		})
		if err != nil { panic(err) }
	}
}

func (f BasicField) WriteUnmarshal(buf *bytes.Buffer) {
	if shouldSkipSerdes(f.Tag) { return }

	cast := tagSearchCast(f.Tag)
	debugPrintln("Cast: ", cast)

	apiType := f.Type
	if cast != "" {
		// For unmarshal, we reverse the cast with the underlying type, because we need to decode the casted type then cast it to the underlying type
		apiType = cast
		cast = f.Type
	}

	apiName, supported := supportedApis[apiType]
	if supported {
		err := BasicTemp.ExecuteTemplate(buf, "basic_unmarshal", map[string]any{
			"Name": f.Name,
			"ApiName": apiName,
			"Type": apiType,
			"Cast": cast,
			// "Type": f.GetType(),
			// "Cast": cast,
		})
		if err != nil { panic(err) }
	} else {
		// debugPrintln("Found Struct: ", f.Name)
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
	// 	debugPrintln("Couldn't find type, assuming its a struct: ", f.Name)
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
	Len string
	Tag string
	IndexDepth int
}

func (f *ArrayField) SetName(name string) {
	f.Name = name
}
func (f *ArrayField) SetTag(tag string) {
	f.Tag = tag
	f.Field.SetTag(tag)
}
func (f *ArrayField) GetType() string {
	return fmt.Sprintf("[%s]%s", f.Len, f.Field.GetType())
}

func (f ArrayField) WriteEquality(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	f.Field.WriteEquality(innerBuf)


	err := BasicTemp.ExecuteTemplate(buf, "array_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"Index": fmt.Sprintf("i%d", f.IndexDepth),
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}

func (f ArrayField) WriteMarshal(buf *bytes.Buffer) {
	if shouldSkipSerdes(f.Tag) { return }

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
	if shouldSkipSerdes(f.Tag) { return }

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
	f.Field.SetTag(tag)
}
func (f *SliceField) GetType() string {
	return fmt.Sprintf("[]%s", f.Field.GetType())
}

func (f SliceField) WriteEquality(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)
	idxVar := fmt.Sprintf("i%d", f.IndexDepth)
	f.Field.SetName(fmt.Sprintf("%s[%s]", f.Name, idxVar))
	f.Field.WriteEquality(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "slice_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"Type": f.Field.GetType(),
		"Index": idxVar,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}

func (f SliceField) WriteMarshal(buf *bytes.Buffer) {
	if shouldSkipSerdes(f.Tag) { return }

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
	if shouldSkipSerdes(f.Tag) { return }

	innerBuf := new(bytes.Buffer)
	varName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(varName)
	f.Field.WriteUnmarshal(innerBuf)

	// debugPrintln("GETTYPE: ", f.Field.GetType())
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
	f.Key.SetTag(tag)
	f.Val.SetTag(tag)
}
func (f *MapField) GetType() string {
	return fmt.Sprintf("map[%s]%s", f.Key.GetType(), f.Val.GetType())
}

func (f MapField) WriteEquality(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	keyIdxName := fmt.Sprintf("k%d", f.IndexDepth)
	// f.Key.SetName(keyIdxName)
	// f.Key.WriteEquality(innerBuf)

	valIdxName := fmt.Sprintf("v%d", f.IndexDepth)
	f.Val.SetName(valIdxName)
	f.Val.WriteEquality(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "map_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"KeyIdx": keyIdxName,
		"ValIdx": valIdxName,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}

func (f MapField) WriteMarshal(buf *bytes.Buffer) {
	if shouldSkipSerdes(f.Tag) { return }

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
	if shouldSkipSerdes(f.Tag) { return }

	innerBuf := new(bytes.Buffer)
	keyVarName := fmt.Sprintf("key%d", f.IndexDepth)
	f.Key.SetName(keyVarName)
	f.Key.WriteUnmarshal(innerBuf)

	valVarName := fmt.Sprintf("val%d", f.IndexDepth)
	f.Val.SetName(valVarName)
	f.Val.WriteUnmarshal(innerBuf)

	// debugPrintln("GETTYPE: ", f.GetType(), f.Key.GetType(), f.Val.GetType())
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
	f.Field.SetTag(tag)
}
func (f *AliasField) GetType() string {
	return fmt.Sprintf("%s", f.Field.GetType())
}

func (f AliasField) WriteEquality(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteEquality(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "alias_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"AliasType": f.Name,
		"Type": f.GetType(),
		"ValName": valName,
		"InnerCode": string(innerBuf.Bytes()),
	})
	if err != nil { panic(err) }
}

func (f AliasField) WriteMarshal(buf *bytes.Buffer) {
	if shouldSkipSerdes(f.Tag) { return }

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
	if shouldSkipSerdes(f.Tag) { return }

	innerBuf := new(bytes.Buffer)
	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteUnmarshal(innerBuf)

	// debugPrintln("ALIAS_GETTYPE: ", f.GetType(), f.Field.GetType())
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

type UnionField struct {
	Name string
	UnionTag int // This is the actual ID used to tag the data in the union
	Tag string // This is the tag string after a specific field
	Field Field
	IndexDepth int
}

func (f *UnionField) SetName(name string) {
	f.Name = name
}
func (f *UnionField) SetTag(tag string) {
	f.Tag = tag
	f.Field.SetTag(tag)
}
func (f *UnionField) GetType() string {
	return f.Field.GetType()
}

func (f UnionField) WriteEquality(buf *bytes.Buffer) {
	err := BasicTemp.ExecuteTemplate(buf, "union_case_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"Type": f.GetType(),
		"Tag": f.UnionTag,
	})
	if err != nil { panic(err) }
}

//TODO: you could probably support basic types by just marshalling the f.Field code and putting it in the union case statement
func (f UnionField) WriteMarshal(buf *bytes.Buffer) {
	err := BasicTemp.ExecuteTemplate(buf, "union_case_marshal", map[string]any{
		"Name": f.Name,
		"Type": f.GetType(),
		"Tag": f.UnionTag,
	})
	if err != nil { panic(err) }
}


func (f UnionField) WriteUnmarshal(buf *bytes.Buffer) {
	// debugPrintln("ALIAS_GETTYPE: ", f.GetType(), f.Field.GetType())
	err := BasicTemp.ExecuteTemplate(buf, "union_case_unmarshal", map[string]any{
		"Name": f.Name,
		"Type": f.GetType(),
		"Tag": f.UnionTag,
	})
	if err != nil { panic(err) }
}

type PointerField struct {
	Name string
	Field Field
	IndexDepth int
}

func (f *PointerField) SetName(name string) {
	f.Name = name
}
func (f *PointerField) SetTag(tag string) {
	f.Field.SetTag(tag)
}
func (f *PointerField) GetType() string {
	return f.Field.GetType()
}

func (f PointerField) WriteEquality(buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteEquality(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "pointer_equality", map[string]any{
		"Name": f.Name,
		"Name2": "t"+f.Name,
		"Type": f.GetType(),
		"ValName": valName,
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

//TODO: you could probably support basic types by just marshalling the f.Field code and putting it in the union case statement
func (f PointerField) WriteMarshal(buf *bytes.Buffer) {
	// TODO: f.Field has the tag
	// if shouldSkipSerdes(f.Tag) { return }

	innerBuf := new(bytes.Buffer)

	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteMarshal(innerBuf)

	err := BasicTemp.ExecuteTemplate(buf, "pointer_marshal", map[string]any{
		"Name": f.Name,
		"Type": f.GetType(),
		"ValName": valName,
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}


func (f PointerField) WriteUnmarshal(buf *bytes.Buffer) {
	// TODO: f.Field has the tag
	// if shouldSkipSerdes(f.Tag) { return }

	innerBuf := new(bytes.Buffer)
	valName := fmt.Sprintf("value%d", f.IndexDepth)
	f.Field.SetName(valName)
	f.Field.WriteUnmarshal(innerBuf)

	// debugPrintln("ALIAS_GETTYPE: ", f.GetType(), f.Field.GetType())
	err := BasicTemp.ExecuteTemplate(buf, "pointer_unmarshal", map[string]any{
		"Name": f.Name,
		"Type": f.GetType(),
		"ValName": valName,
		"ValType": f.Field.GetType(),
		"InnerCode": innerBuf.String(),
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
	DirectiveUnionDef
)


func (v *Visitor) Output(filename string) {
	if len(v.requests) == 0 && len(v.structs2) == 0 && len(v.structs) == 0 {
		fmt.Println("Skipping: No tagged structs:", filename)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("package " + v.pkg.Name)

	if len(v.usedImports) > 0 {
		buf.WriteString(`
import (
`)


		// 	buf.WriteString(`
		// import (
		// 	"github.com/unitoftime/cod/backend"
		// `)

		toSort := make([]string, 0)
		for k := range v.usedImports {
			toSort = append(toSort, k)
		}
		sort.Strings(toSort)

		for _, k := range toSort {
			debugPrintln("Used Import: ", k)
			path, ok := v.imports[k]
			debugPrintln("Used Import: ", k, path, ok)
			if !ok {
				// fmt.Printf("%s: couldnt find import: %s\n", path, k)
				panic(fmt.Sprintf("%s: couldnt find import: %s", path, k))
			}
			buf.WriteString("\n"+path+"\n")
		}
		buf.WriteString(`
)`)
	}

	toSort := make([]string, 0)
	for k := range v.structs {
		toSort = append(toSort, k)
	}
	sort.Strings(toSort)

	marshBuf := bytes.NewBuffer([]byte{})
	unmarshBuf := bytes.NewBuffer([]byte{})
	for _, k := range toSort {
		sd, _ := v.structs[k]
		if sd.Directive == DirectiveNone { continue }
		if sd.Directive == DirectiveUnionDef { continue }

		marshBuf.Reset()
		unmarshBuf.Reset()

		debugPrintln("Struct: ", sd.Name)

		// If no fields, then its a blank struct
		if len(sd.Fields) <= 0 {
			// Write the encode func
			err := BasicTemp.ExecuteTemplate(buf, "blank_marshal_func", map[string]any{
				"Name": sd.Name,
			})
			if err != nil { panic(err) }

			// Write the decode func
			err = BasicTemp.ExecuteTemplate(buf, "blank_unmarshal_func", map[string]any{
				"Name": sd.Name,
			})
			if err != nil { panic(err) }

			err = BasicTemp.ExecuteTemplate(buf, "blank_equality_func", map[string]any{
				"Name": sd.Name,
			})
			if err != nil { panic(err) }
			continue
		}

		// Write the marshal code
		v.WriteStructMarshal(&sd, marshBuf)
		// Write the unmarshal code
		v.WriteStructUnmarshal(&sd, unmarshBuf)

		// Write the encode func
		err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
			"Name": sd.Name,
			"MarshalCode": string(marshBuf.Bytes()),
		})
		if err != nil { panic(err) }

		// Write the decode func
		err = BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
			"Name": sd.Name,
			"MarshalCode": string(unmarshBuf.Bytes()),
		})
		if err != nil { panic(err) }

		// Special Union funcs
		v.WriteUnionCodeToBuffer(&sd, buf)

		// Special Equality operator
		v.WriteEqualityCodeToBuffer(&sd, buf)
	}

	for _, k := range toSort {
		sd, _ := v.structs[k]
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

	for _, k := range toSort {
		sd, _ := v.structs[k]
		for _, req := range v.requests[k] {
			switch req.Type {
			case RequestTypeComponent:
				err := BasicTemp.ExecuteTemplate(buf, "ecs_component", map[string]any{
					"Name": sd.Name,
				})
				if err != nil { panic(err) }
			}
		}
	}

	outputFile(filename, buf)
}
