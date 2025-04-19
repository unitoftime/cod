package main

import "bytes"

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

func GenerateSerdesData(sd StructData, buf *bytes.Buffer) {
	debugPrintln("Struct: ", sd.Name)

	// If no fields, then its a blank struct
	if len(sd.Fields) <= 0 {
		GenerateBlankSerdesData(sd, buf)
		return
	}

	WriteStructMarshal(sd, buf)
	WriteStructUnmarshal(sd, buf)
	WriteStructEquality(sd, buf)
}

func GenerateBlankSerdesData(sd StructData, buf *bytes.Buffer) {
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
}

func WriteStructMarshal(sd StructData, buf *bytes.Buffer) {
	marshBuf := new(bytes.Buffer)
	for _, f := range sd.Fields {
		f.WriteMarshal(marshBuf)
	}
	err := BasicTemp.ExecuteTemplate(buf, "marshal_func", map[string]any{
		"Name": sd.Name,
		"MarshalCode": marshBuf.String(),
	})
	if err != nil { panic(err) }
}

func WriteStructUnmarshal(sd StructData, buf *bytes.Buffer) {
	unmarshBuf := new(bytes.Buffer)
	for _, f := range sd.Fields {
		f.WriteUnmarshal(unmarshBuf)
	}
	// Write the decode func
	err := BasicTemp.ExecuteTemplate(buf, "unmarshal_func", map[string]any{
		"Name": sd.Name,
		"MarshalCode": unmarshBuf.String(),
	})
	if err != nil { panic(err) }
}

func WriteStructEquality(s StructData, buf *bytes.Buffer) {
	innerBuf := new(bytes.Buffer)

	for _, f := range s.Fields {
		f.WriteEquality(innerBuf)
	}
	// Write the equality func
	err := BasicTemp.ExecuteTemplate(buf, "equality_func", map[string]any{
		"Name": s.Name,
		"InnerCode": innerBuf.String(),
	})
	if err != nil { panic(err) }
}
