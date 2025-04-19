package main

import "bytes"

func GenerateUnionData(sd StructData, csv []string, structs map[string]StructData, buf *bytes.Buffer) {
	marshBuf := new(bytes.Buffer)
	unmarshBuf := new(bytes.Buffer)

	debugPrintln("Union: ", sd.Name)

	// If no fields, then its a blank struct
	if len(sd.Fields) <= 0 {
		// Write the encode func
		GenerateBlankSerdesData(sd, buf)
		return
	}

	// Write the marshal code
	WriteUnionMarshal(sd, csv, structs, marshBuf)
	err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
		"Name": sd.Name,
		"MarshalCode": marshBuf.String(),
	})
	if err != nil { panic(err) }

	// Write the unmarshal code
	WriteUnionUnmarshal(sd, csv, structs, unmarshBuf)
	err = BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
		"Name": sd.Name,
		"MarshalCode": unmarshBuf.String(),
	})
	if err != nil { panic(err) }

	// Special Union funcs
	WriteUnionCodeToBuffer(sd, csv, structs, buf)
	WriteUnionEqualityCode(sd, csv, structs, buf)

	//----------------------------------------
	// - Union Helper Functions
	//----------------------------------------

	// Create constructors, getters, setters per union type
	err = BasicTemp.ExecuteTemplate(buf, "union_getter", map[string]any{
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

func WriteUnionMarshal(sd StructData, csv []string, structs map[string]StructData, buf *bytes.Buffer) {
	// For unions we lookup the union def which must be the first csv element
	unionDefName := csv[0]
	unionDef, ok := structs[unionDefName]
	if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

	innerBuf := new(bytes.Buffer)
	debugPrintln("UnionDef: ", unionDef.Fields)
	for i, f := range unionDef.Fields {
		f := NewUnionField(f, i+1)
		f.WriteMarshal(innerBuf)
	}
	err := BasicTemp.ExecuteTemplate(buf, "union_marshal", map[string]any{
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

func WriteUnionEqualityCode(sd StructData, csv []string, structs map[string]StructData, buf *bytes.Buffer) {
	// For unions we lookup the union def which must be the first csv element
	unionDefName := csv[0]
	unionDef, ok := structs[unionDefName]
	if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

	innerBuf := new(bytes.Buffer)
	for i, f := range unionDef.Fields {
		f := NewUnionField(f, i+1)
		err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_equality", map[string]any{
			"Name": f.Name,
			"Name2": "t"+f.Name,
			"Type": f.GetType(),
			"Tag": f.UnionTag,
		})
		if err != nil { panic(err) }
	}

	err := BasicTemp.ExecuteTemplate(buf, "union_equality_func", map[string]any{
		"Name": sd.Name,
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

func WriteUnionUnmarshal(sd StructData, csv []string, structs map[string]StructData, buf *bytes.Buffer) {
	// For unions we lookup the union def which must be the first csv element
	unionDefName := csv[0]
	unionDef, ok := structs[unionDefName]
	if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

	innerBuf := new(bytes.Buffer)
	for i, f := range unionDef.Fields {
		f = &UnionField{
			Name: f.GetName(),
			UnionTag: i+1,
			Field: f,
		}
		f.WriteUnmarshal(innerBuf)
	}

	err := BasicTemp.ExecuteTemplate(buf, "union_unmarshal", map[string]any{
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}

func WriteUnionCodeToBuffer(sd StructData, csv []string, structs map[string]StructData, buf *bytes.Buffer) {
	// For unions we lookup the union def which must be the first csv element
	unionDefName := csv[0]
	unionDef, ok := structs[unionDefName]
	if !ok { panic("Union def must be first element: //cod:union <UnionDefType>") }

	debugPrintln("UDEF: ", unionDef.Fields)
	// GetTag()
	{
		innerBuf := new(bytes.Buffer)
		for i, f := range unionDef.Fields {
			f := NewUnionField(f, i+1)
			err := BasicTemp.ExecuteTemplate(innerBuf, "union_case_get_tag", map[string]any{
				"Name": f.Name,
				"Type": f.GetType(),
				"Tag": f.UnionTag,
			})
			if err != nil { panic(err) }
		}

		err := BasicTemp.ExecuteTemplate(buf, "union_get_tag_func", map[string]any{
			"Name": sd.Name,
			"InnerCode": innerBuf.String(),
		})
		if err != nil { panic(err) }
	}

	// GetSize()
	{
		err := BasicTemp.ExecuteTemplate(buf, "union_get_size_func", map[string]any{
			"Name": sd.Name,
			"Size": len(unionDef.Fields) + 1, // Note: + 1 b/c 0 is the nil case
		})
		if err != nil { panic(err) }
	}
}
