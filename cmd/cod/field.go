package main

import (
	"bytes"
	"fmt"
)

// TODO: Maybe?
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


type Field interface {
	SetTag(string)
	GetName() string
	SetName(string)
	GetType() string

	// TODO: Remove these
	WriteEquality(*bytes.Buffer)
	WriteMarshal(*bytes.Buffer)
	WriteUnmarshal(*bytes.Buffer)
}

type BasicField struct {
	Name string
	Type string
	Tag string
}

func (f *BasicField) GetName() string {
	return f.Name
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

func (f *ArrayField) GetName() string {
	return f.Name
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

func (f *SliceField) GetName() string {
	return f.Name
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

func (f *MapField) GetName() string {
	return f.Name
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

func (f *AliasField) GetName() string {
	return f.Name
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

func NewUnionField(field Field, tag int) UnionField {
	return UnionField{
		Name: field.GetName(), // TODO: Is this even needed?
		UnionTag: tag,
		Field: field,
	}
}
type UnionField struct {
	Name string // TODO: Is this even needed?
	UnionTag int // This is the actual ID used to tag the data in the union
	Tag string // This is the tag string after a specific field
	Field Field
	// IndexDepth int
}

func (f *UnionField) GetName() string {
	return f.Name
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

func (f *PointerField) GetName() string {
	return f.Name
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
