package main

import (
	"bytes"
	"fmt"
	"go/types"
	"maps"
	"slices"
	"sort"
	"time"

	"golang.org/x/tools/go/packages"
)

type generator struct {
	buf *bytes.Buffer

	imports map[string]string // Maps a selector source to a package path
	usedImports map[string]struct{} // List of used package selectors
}

func (g *generator) generateEcsComponentCod(pkg *packages.Package, typeDef types.Object, csv []string) {
	g.imports["ecs"] = "\"github.com/unitoftime/ecs\""
	g.usedImports["ecs"] = struct{}{}

	err := BasicTemp.ExecuteTemplate(g.buf, "ecs_component", map[string]any{
		"Name": typeDef.Name(),
	})
	if err != nil { panic(err) }
}

func newMain() {
	now := time.Now()
	defer printDuration("newMain", now)

	// TODO: Recursively search through package root? Probably faster than having a generate at each layer

	requests := findTaggedObjects()
	// fmt.Printf("Tags: %+v\n", requests)

	generator := generator{
		buf: &bytes.Buffer{},

		imports: make(map[string]string),
		usedImports: make(map[string]struct{}),
	}

	cfg := &packages.Config{Mode: packages.LoadTypes | packages.LoadSyntax | packages.LoadImports}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		panic(err)
	}

	// TODO: Not 100% sure what this does
	// packages.PrintErrors(pkgs)

	pkgName := ""
	for _, pkg := range pkgs {
		pkgName = pkg.Name

		defs := make(map[string]types.Object)
		for _, typeDef := range pkg.TypesInfo.Defs {
			if typeDef == nil { continue }
			_, ok := requests[typeDef.Name()]
			if !ok { continue }
			defs[typeDef.Name()] = typeDef
		}

		// TODO: Either store a buffer per def and sort them then output at the end, or sort this map iteration
		// for _, typeDef := range pkg.TypesInfo.Defs {
		for _, typeName := range slices.Sorted(maps.Keys(requests)) {
			typeDef, ok := defs[typeName]
			if !ok { continue }
			reqs, ok := requests[typeDef.Name()]
			if !ok { continue }
			// fmt.Println("Requested: ", typeDef.Name(), reqs)

			for _, r := range reqs {
				switch r.Type {
				case RequestTypeComponent:
					generator.generateEcsComponentCod(pkg, typeDef, r.CSV)
				// case RequestTypeEnum:
				// 	generator.generateUnsafeEnumCode(pkg, typeDef, r.CSV)
				// case RequestTypeEnumSafe:
				// 	generator.generateSafeEnumCode(pkg, typeDef, r.CSV)
				// // case RequestTypeStructLegacy:
				// // 	generator.generateStructEqualityCode(pkg, typeDef)
				}
			}
		}
	}

	var finalBuf bytes.Buffer
	fmt.Fprintf(&finalBuf, "package %s", pkgName)
	generator.writeImports(&finalBuf)

	fmt.Fprint(&finalBuf, generator.buf.String())

	outputFile("cod.gen.go", &finalBuf)
}

//--------------------------------------------------------------------------------

func (v *generator) writeImports(buf *bytes.Buffer) {
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
}
