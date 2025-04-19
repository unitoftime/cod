package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func findTaggedObjects() map[string][]GenRequest {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	retRequests := make(map[string][]GenRequest)
	for _, pkg := range packages {
		bv := &findVisitor{
			fset: fset,
			requests: make(map[string][]GenRequest),
		}

		// We start walking our Visitor `bv` through the AST in a depth-first way.
		ast.Walk(bv, pkg)

		// Merge maps
		for k, v := range bv.requests {
			retRequests[k] = append(retRequests[k], v...)
		}

	}
	return retRequests
}

type findVisitor struct {
	fset *token.FileSet // The fileset of the package we are processing
	file *ast.File // The file we are currently processing (Can be nil if we haven't started processing a file yet!)
	cmap ast.CommentMap // The comment map of the file we are processing

	requests map[string][]GenRequest
}
func (v *findVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil { return nil }

	// If we are a package, then just keep searching
	_, ok := node.(*ast.Package)
	if ok { return v }

	// If we are a file, then store some data in the visitor so we can use it later
	file, ok := node.(*ast.File)
	if ok {
		v.file = file
		v.cmap = ast.NewCommentMap(v.fset, file, file.Comments)

		// for _, importSpec := range file.Imports {
		// 	debugPrintln("IMPORT SPEC: ", importSpec)
		// 	path := importSpec.Path.Value

		// 	name := strings.TrimSuffix(filepath.Base(path), `"`)

		// 	// If there was a custom name, use that
		// 	nameIdent := importSpec.Name
		// 	if nameIdent != nil {
		// 		name = nameIdent.Name
		// 	}
		// 	debugPrintln("IMPORT: ", name, path)
		// 	v.imports[name] = path
		// }
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
		v.getRequests(*gen, cgroups)

		return nil
	}

	// If all else fails, then keep looking
	return v
}

func (v *findVisitor) getRequests(decl ast.GenDecl, cGroups []*ast.CommentGroup) {
	// Skip everything that isn't a type. we can only generate for types
	if decl.Tok != token.TYPE {
		decl.Doc = nil // nil the Doc field so that we don't print it
		return
	}

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			v.checkComments(s.Name.Name, decl)
		}
	}
}

type directiveHandler struct {
	str string
	RequestType RequestType
}
var directiveSearch = []directiveHandler{
	{"//cod:component", RequestTypeComponent},
	// {"//cod:enum", RequestTypeEnum},
	// {"//cod:safe-enum", RequestTypeEnumSafe},
	// {"//cod:struct", RequestTypeStructLegacy},
}

func (v *findVisitor) checkComments(name string, t ast.GenDecl) {
	if t.Doc == nil {
		return
	}

	for _, c := range t.Doc.List {
		for _, search := range directiveSearch {
			after, found := strings.CutPrefix(c.Text, search.str)
			if found {
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
}
