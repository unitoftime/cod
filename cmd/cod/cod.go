package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"go/ast"
	"go/parser"

	"go/token"

	"path/filepath"
)

var skip = flag.String("skip", ".git,.github", "directories to match and skip")
var verbose = flag.Bool("v", false, "print more output")

func main() {
	now := time.Now()
	defer printDuration("cod generate time", now)

	flag.Parse()

	skipMap := make(map[string]struct{})
	if skip != nil {
		skipList := strings.Split(*skip, ",")
		for i := range skipList {
			skipMap[strings.TrimSpace(skipList[i])] = struct{}{}
		}
	}

	generateAll(".", skipMap)
}

func generateAll(dir string, skipMap map[string]struct{}) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() { return nil } // Skip if not the directory
		_, skip := skipMap[d.Name()]
		if skip { return filepath.SkipDir }

		fullPath := filepath.Join(dir, path)
		generatePackage(fullPath)
		return nil
	})
}

func generatePackage(dir string) {
	if *verbose {
		fmt.Println("Parsing Directory:", dir)
	}

	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range packages {
		// fmt.Println("Parsing Package:", pkg.Name)
		bv := &Visitor{
			pkg: pkg,
			fset: fset,
			requests: make(map[string][]GenRequest),

			structs: make(map[string]StructData),
			imports: make(map[string]string),
			usedImports: make(map[string]bool),
		}

		// Register some common imports in case they are needed
		bv.imports["backend"] = "\"github.com/unitoftime/cod/backend\""
		bv.imports["fmt"] = "\"fmt\""


		// We start walking our Visitor `bv` through the AST in a depth-first way.
		ast.Walk(bv, pkg)

		bv.Output(filepath.Join(dir, "cod_encode.go"))
	}
}
func (v *Visitor) formatGen(decl ast.GenDecl) (StructData, bool) {
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
			add, trackImports := v.getRequests(s.Name.Name, decl)

			if !add {
				return structData, false // Skip because it wasn't marked
			}

			debugPrintln("TypeSpec: ", s.Name.Name)
			debugPrintf("TypeSpec: %T\n", s.Type)
			structData.Name = s.Name.Name

			// debugPrintf("Struct Type: %T\n", s.Type)
			sType, ok := s.Type.(*ast.StructType)
			if !ok {
				// Not a struct, then its an alias. So handle that if we can
				name := "t"
				idxDepth := 0
				field := v.generateField(name, idxDepth+1, s.Type, trackImports)
				fields = append(fields, &AliasField{
					Name: name,
					AliasType: s.Name.Name,
					Field: field,
					IndexDepth: idxDepth,
				})
				continue
			}

			debugPrintln("Fields: ", sType.Fields.List)

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
					field := v.generateField("t." + name, idxDepth+1, f.Type, trackImports)
					if f.Tag != nil {
						field.SetTag(f.Tag.Value)
					}

					fields = append(fields, field)
				} else {
					for _, n := range f.Names {
						debugPrintln("Field: ", n.Name, f.Type, f.Tag)
						debugPrintf("%T\n", f.Type)

						idxDepth := 0
						field := v.generateField("t." + n.Name, idxDepth+1, f.Type, trackImports)
						if f.Tag != nil {
							field.SetTag(f.Tag.Value)
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

func (v *Visitor) generateField(name string, idxDepth int, node ast.Node, trackImports bool) Field {
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
		field := v.generateField(name, idxDepth+1, expr.X, trackImports)
		return &PointerField{
			Name: name,
			Field: field,
			IndexDepth: idxDepth,
		}

	case *ast.ArrayType:
		// debugPrintf("ARRAY %T %T\n", expr.Len, expr.Elt)

		if expr.Len == nil {
			idxString := fmt.Sprintf("[i%d]", idxDepth)
			field := v.generateField(name + idxString, idxDepth + 1, expr.Elt, trackImports)
			return &SliceField{
				Name: name,
				// Type: field.GetType(),
				Field: field,
				IndexDepth: idxDepth,
			}
		} else {
			idxString := fmt.Sprintf("[i%d]", idxDepth)
			field := v.generateField(name + idxString, idxDepth + 1, expr.Elt, trackImports)

			lengthIdentName := ""
			switch lenT := expr.Len.(type) {
			case *ast.Ident:
				lengthIdentName = lenT.Name
			case *ast.BasicLit:
				lengthIdentName = lenT.Value
			default:
				// panic(fmt.Sprintf("unhandled array length type: %T", expr.Len))
				lengthIdentName = "???"
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
		key := v.generateField(name + keyString, idxDepth + 1, expr.Key, trackImports)
		val := v.generateField(name + valString, idxDepth + 1, expr.Value, trackImports)
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

		if trackImports {
			v.usedImports[x.Name] = true // Store the import name so we can pull it later
		}

		return field

		// TODO: I made these all invalid because they cant be used for serialization, but because you generate ecs code from here you need to handle them. so I just put a blank basic field
	case *ast.FuncType:
		return &BasicField{} // Invalid
	case *ast.IndexExpr:
		return &BasicField{} // Invalid
	case *ast.IndexListExpr:
		return &BasicField{} // Invalid
	case *ast.StructType:
		return &BasicField{} // Invalid
	case *ast.ChanType:
		return &BasicField{} // Invalid
	default:
		panic(fmt.Sprintf("%s:, unknown type %T", name, expr))
	}

	return nil
}

type directiveHandler struct {
	str string
	RequestType RequestType

	// TODO: Ideally the required imports would be tracked by the generating code because doing it up here is super broad and doesnt handle perfectly
	RequiredImports []string
	TrackImports bool // If true, this requestType means we need to track and use the imports of the inner types in the struct
}
var directiveSearch = []directiveHandler{
	{"//cod:component", RequestTypeComponent, []string{"ecs"}, false},
	{"//cod:struct", RequestTypeSerdes, []string{"backend"}, true},
	{"//cod:union", RequestTypeUnion, []string{"fmt"}, true},
	{"//cod:def", RequestTypeUnionDef, []string{}, true},

	// {"//cod:enum", RequestTypeEnum},
	// {"//cod:safe-enum", RequestTypeEnumSafe},
	// {"//cod:struct", RequestTypeStructLegacy},
}

func (v *Visitor) getRequests(name string, t ast.GenDecl) (bool, bool) {
	if t.Doc == nil {
		return false, false
	}

	generatedCodeRequiresImports := false
	for _, c := range t.Doc.List {
		for _, search := range directiveSearch {
			after, found := strings.CutPrefix(c.Text, search.str)
			if found {
				if search.TrackImports {
					generatedCodeRequiresImports = search.TrackImports
				}

				for _, reqImport := range search.RequiredImports {
					v.usedImports[reqImport] = true
				}

				csv := strings.Split(after, ",")
				for i := range csv {
					csv[i] = strings.TrimSpace(csv[i])
				}

				v.requests[name] = append(v.requests[name], GenRequest{
					Type: search.RequestType,
					CSV: csv,
				})
			}
		}
	}
	return (len(v.requests[name]) > 0), generatedCodeRequiresImports
}

type RequestType int
const (
	RequestTypeComponent RequestType = iota
	RequestTypeSerdes
	RequestTypeUnion
	RequestTypeUnionDef
)
type GenRequest struct {
	Type RequestType
	CSV []string
}

type Visitor struct {
	pkg *ast.Package  // The package that we are processing
	fset *token.FileSet // The fileset of the package we are processing
	file *ast.File // The file we are currently processing (Can be nil if we haven't started processing a file yet!)
	cmap ast.CommentMap // The comment map of the file we are processing

	requests map[string][]GenRequest

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
		sd, ok := v.formatGen(*gen)
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
	Fields []Field
}

func (v *Visitor) Output(filename string) {
	if len(v.requests) == 0 && len(v.structs) == 0 {
		return // Skip: no tagged structs
	}

	buf := new(bytes.Buffer)
	toSort := make([]string, 0)
	for k := range v.structs {
		toSort = append(toSort, k)
	}
	sort.Strings(toSort)

	// Generate all of the requests
	for _, k := range toSort {
		sd, _ := v.structs[k]
		for _, req := range v.requests[k] {
			switch req.Type {
			case RequestTypeComponent:
				err := BasicTemp.ExecuteTemplate(buf, "ecs_component", map[string]any{
					"Name": sd.Name,
				})
				if err != nil { panic(err) }

			case RequestTypeSerdes:
				GenerateSerdesData(sd, buf)
			case 	RequestTypeUnion:
				GenerateUnionData(sd, req.CSV, v.structs, buf)
			case RequestTypeUnionDef:
				// Noop: We only have a request for this one because we need to look it up from from actual union code gen
			}
		}
	}

	fileBuf := new(bytes.Buffer)
	fileBuf.WriteString("package " + v.pkg.Name)
	v.WriteImports(fileBuf)
	fileBuf.Write(buf.Bytes())

	outputFile(filename, fileBuf)
}

func (v *Visitor) WriteImports(buf *bytes.Buffer) {
	if len(v.usedImports) > 0 {
		buf.WriteString(`
import (
`)

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
}
