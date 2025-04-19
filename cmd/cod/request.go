package main

import (
	"go/ast"
	"strings"
)

type RequestType int

type GenRequest struct {
	Type RequestType
	CSV []string
}

type requestConfig struct {
	str string
	RequestType RequestType

	// TODO: Ideally the required imports would be tracked by the generating code because doing it up here is super broad and doesnt handle perfectly
	RequiredImports []string
	TrackImports bool // If true, this requestType means we need to track and use the imports of the inner types in the struct
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
