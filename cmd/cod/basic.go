package main

import (
	// "bytes"
	"text/template"
)

var BasicTemp *template.Template

func addTemplate(name string, dat string) {
	template.Must(BasicTemp.New(name).Parse(dat))
}

func addStructTemplate(name string, dat string) {
	template.Must(BasicTemp.New(name).Parse(dat))
}

func init() {
	BasicTemp = template.New("BasicTemp")

	// Marshal/Unmarshal Functions
	addTemplate("marshal_func", `
func (t {{.Name}})EncodeCod(bs []byte) []byte {
{{.MarshalCode}}
return bs
}
`)

	addTemplate("unmarshal_func", `
func (t *{{.Name}})DecodeCod(bs []byte) (int, error) {
var err error
var n int
var nOff int

{{.MarshalCode}}

// println("{{.Name}}:", n)
return n, err
}
`)

	// Blank Marshal/Unmarshal Functions
	addTemplate("blank_marshal_func", `
func (t {{.Name}})EncodeCod(bs []byte) []byte {
return bs
}
`)

	addTemplate("blank_unmarshal_func", `
func (t *{{.Name}})DecodeCod(bs []byte) (n int, err error) {
return
}
`)

	// Standard Types
	addTemplate("basic_marshal", `
bs = backend.Write{{.ApiName}}(bs, {{.Cast}}({{.Name}}))
`)

// {{.Name}}, nOff, err = backend.Read{{.ApiName}}(bs[n:])
// if err != nil { return 0, err }
// n += nOff
	addTemplate("basic_unmarshal", `
{
var decoded {{.Type}}
decoded, nOff, err = backend.Read{{.ApiName}}(bs[n:])
if err != nil { return 0, err }
n += nOff
{{.Name}} = {{.Cast}}(decoded)
}
`)


	// Struct
	addTemplate("struct_marshal", `
bs = {{.Name}}.EncodeCod(bs)`)

	addTemplate("struct_unmarshal", `
{
var decoded {{.Type}}
nOff, err = decoded.DecodeCod(bs[n:])
if err != nil { return 0, err }
n += nOff
{{.Name}} = decoded
}
`)

	// TODO: could also unroll the loop here?
	// Arrays
	addTemplate("array_marshal", `
for {{.Index}} := range {{.Name}} {
   {{.InnerCode}}
}`)

	addTemplate("array_unmarshal", `
for {{.Index}} := range {{.Name}} {
   {{.InnerCode}}

   if err != nil {
      return 0, err
   }
}`)

	// Slice
	addTemplate("slice_marshal", `
{
bs = backend.WriteVarUint64(bs, uint64(len({{.Name}})))
for {{.Index}} := range {{.Name}} {
   {{.InnerCode}}
}
}`)

	addTemplate("slice_unmarshal", `
{
  var length uint64
	length, nOff, err = backend.ReadVarUint64(bs[n:])
	if err != nil { return 0, err }
  n += nOff

for {{.Index}} := 0; {{.Index}} < int(length); {{.Index}}++ {
   var {{.VarName}} {{.Type}}
   {{.InnerCode}}
   if err != nil {
      return 0, err
   }

   {{.Name}} = append({{.Name}}, {{.VarName}})
}
}`)


	// Map
	addTemplate("map_marshal", `
{
bs = backend.WriteVarUint64(bs, uint64(len({{.Name}})))

for {{.KeyIdx}}, {{.ValIdx}} := range {{.Name}} {
   {{.InnerCode}}
}

}`)

	addTemplate("map_unmarshal", `
{
  var length uint64
	length, nOff, err = backend.ReadVarUint64(bs[n:])
	if err != nil { return 0, err }
  n += nOff

if {{.Name}} == nil {
{{.Name}} = make({{.Type}})
}

for {{.Index}} := 0; {{.Index}} < int(length); {{.Index}}++ {
   var {{.KeyVar}} {{.KeyType}}
   var {{.ValVar}} {{.ValType}}

   {{.InnerCode}}
   if err != nil {
      return 0, err
   }

   {{.Name}}[{{.KeyVar}}] = {{.ValVar}}
}
}`)


	// Alias
	addTemplate("alias_marshal", `
{
   {{.ValName}} := {{.Type}}({{.Name}})
   {{.InnerCode}}

}`)

	addTemplate("alias_unmarshal", `
{
   var {{.ValName}} {{.Type}}
   {{.InnerCode}}
   *{{.Name}} = {{.AliasType}}({{.ValName}})
}`)

	// --------------------------------------------------------------------------------
	// Union
	// --------------------------------------------------------------------------------
	// TODO: Should I do uvarints for tags? 255 is an absolutely massive union...
	addTemplate("union_get_tag_func", `
func (t {{.Name}})Tag() uint8 {
   rawVal := t.Get()
   if rawVal == nil {
      // Zero tag indicates nil
      return 0
   }

   switch rawVal.(type) {
{{.InnerCode}}
   default:
      panic("unknown type placed in union")
   }
}
`)

	// Union cases
	addTemplate("union_case_get_tag", `
   case {{.Type}}:
      return {{.Tag}}
`)

	addTemplate("union_get_size_func", `
func (t {{.Name}})Size() int {
   return {{.Size}}
}
`)

	// TODO: Should I do uvarints for tags? 255 is an absolutely massive union...
	addTemplate("union_marshal", `
   tag := t.Tag()
   bs = backend.WriteUint8(bs, tag)
   if tag == 0 {
      // Zero tag indicates nil, so write nothing else
      return bs
   }

   rawVal := t.Get()
   bs = rawVal.EncodeCod(bs)

   return bs
`)
// Old marshal code. just keeping it around in case
/*
   rawVal := t.Get()
   if rawVal == nil {
      // Zero tag indicates nil
      bs = backend.WriteUint8(bs, 0)
      return bs
   }

   switch sv := rawVal.(type) {
   {{.InnerCode}}
   default:
      panic("unknown type placed in union")
   }

   return bs
*/


	addTemplate("union_unmarshal", `
   var tagVal uint8

   tagVal, nOff, err = backend.ReadUint8(bs[n:])
   if err != nil { return 0, err }
   n += nOff

   switch tagVal {
   case 0: // Zero tag indicates nil
      return 0, nil

   {{.InnerCode}}
   default:
      return 0, backend.ErrUnknownUnionType
   }
`)

	// Union cases
	addTemplate("union_case_marshal", `
   case {{.Type}}:
      bs = backend.WriteUint8(bs, {{.Tag}})
      bs = sv.EncodeCod(bs)
`)

	addTemplate("union_case_unmarshal", `
   case {{.Tag}}:
      var decoded {{.Type}}
      nOff, err = decoded.DecodeCod(bs[n:])
      if err != nil { return 0, err }
      n += nOff

      t.Set(decoded)
`)

	// Union getters and setters
	addTemplate("union_getter", `
func (t {{.Name}}) Get() cod.EncoderDecoder {
   codUnion := cod.Union(t)
   rawVal := codUnion.GetRawValue()
   return rawVal

   // switch rawVal.(type) {
   // {{.InnerCode}}
   // default:
   //    panic("unknown type placed in union")
   // }
}
`)

	addTemplate("union_constructor", `
func New{{.Name}}(v cod.EncoderDecoder) {{.Name}} {
   var ret {{.Name}}
   ret.Set(v)
   return ret
}
`)

// TODO: You could theoretically check and panic if the user passes in an incorrect value
	addTemplate("union_setter", `
func (t *{{.Name}}) Set(v cod.EncoderDecoder) {
   codUnion := cod.Union(*t)
   codUnion.PutRawValue(v)
   *t = {{.Name}}(codUnion)

   // switch tagVal {
   // case 0: // Zero tag indicates nil
   //    return nil

   // {{.InnerCode}}
   // default:
   //    panic("unknown type placed in union")
   // }
   // return err
}
`)

	// Pointer
	addTemplate("pointer_marshal", `
{
   if {{.Name}} == nil {
      // Zero tag indicates nil
      bs = backend.WriteUint8(bs, 0)
   } else {
      bs = backend.WriteUint8(bs, 1)
      {{.ValName}} := *{{.Name}}
      {{.InnerCode}}
   }
}`)

	addTemplate("pointer_unmarshal", `
{
   var tagVal uint8
   tagVal, nOff, err = backend.ReadUint8(bs[n:])
   if err != nil { return 0, err }
   n += nOff

   if tagVal == 0 {
      // Zero tag indicates nil
      {{.Name}} = nil
   } else {
      var {{.ValName}} {{.ValType}}
      {{.InnerCode}}
      {{.Name}} = &{{.ValName}}
   }
}`)


// 	// Struct
// 	addTemplate("reg_struct_marshal", `
// {{.Name}}.EncodeCod(buf)`)
// 	addTemplate("reg_struct_unmarshal", `
// {{.Name}}.DecodeCod(buf)`)

// 	addTemplate("ptr_struct_marshal", `
// {{.Name}}.EncodeCod(buf)`)
// 	addTemplate("ptr_struct_unmarshal", `
// {{.Name}}.DecodeCod(buf)`)


// 	// uint8
// 	addTemplate("reg_uint8_marshal", `
// buf.WriteUint8({{.Name}})`)

// 	addTemplate("reg_uint8_unmarshal", `
// {{.Name}} = buf.ReadUint8()`)

// 	// uint16
// 	addTemplate("reg_uint16_marshal", `
// buf.WriteUint16({{.Name}})`)

// 	addTemplate("reg_uint16_unmarshal", `
// {{.Name}}, err = buf.ReadUint16()`)

// 	// uint32
// 	addTemplate("reg_uint32_marshal", `
// buf.WriteUint32({{.Name}})`)

// 	addTemplate("reg_uint32_unmarshal", `
// {{.Name}}, err = buf.ReadUint32()`)

// 	// uint64
// 	addTemplate("reg_uint64_marshal", `
// buf.WriteUint64({{.Name}})`)

// 	addTemplate("reg_uint64_unmarshal", `
// {{.Name}}, err = buf.ReadUint64()`)

// 	// int32
// 	addTemplate("reg_int32_marshal", `
// buf.WriteInt32({{.Name}})`)

// 	addTemplate("reg_int32_unmarshal", `
// {{.Name}}, err = buf.ReadInt32()`)

// 	// int64
// 	addTemplate("reg_int64_marshal", `
// buf.WriteInt64({{.Name}})`)

// 	addTemplate("reg_int64_unmarshal", `
// {{.Name}}, err = buf.ReadInt64()`)

// 	// float64
// 	addTemplate("reg_float64_marshal", `
// buf.WriteFloat64({{.Name}})`)

// 	addTemplate("reg_float64_unmarshal", `
// {{.Name}} = buf.ReadFloat64()`)

// 	// bool
// 	addTemplate("reg_bool_marshal", `
// buf.WriteBool({{.Name}})`)

// 	addTemplate("reg_bool_unmarshal", `
// {{.Name}} = buf.ReadBool()`)

// 	// string
// 	addTemplate("reg_string_marshal", `
// buf.WriteString({{.Name}})`)

// 	addTemplate("reg_string_unmarshal", `
// {{.Name}}, err = buf.ReadString()`)

// 	// TODO: This is wrong. I need to have alike an "is nil" byte or something. You cant mix up "" with nil
// 	addTemplate("ptr_string_marshal", `
// if {{.Name}} == nil {
//    buf.WriteString("")
// } else {
//    buf.WriteString(*{{.Name}})
// }`)

// 	addTemplate("ptr_string_unmarshal", `
// {{.Name}}, err = buf.ReadString()`)


// 	// Ideas: Generic template

// 	// Standard Types
// 	addTemplate("basic_marshal", `
// buf.Write{{.Type}}({{.Name}})
// `)

// 	addTemplate("basic_unmarshal", `
// {{.Name}}, err = buf.Read{{.Type}}()
// `)

// 	// Struct
// 	addTemplate("struct_marshal", `
// {{.Name}}.EncodeCod(buf)
// `)

// 	addTemplate("struct_unmarshal", `
// err = {{.Name}}.DecodeCod(buf)
// `)

// 	// TODO: could also unroll the loop here?
// 	// Arrays
// 	addTemplate("array_marshal", `
// for {{.Index}} := range {{.Name}} {
//    {{.InnerCode}}
// }`)

// 	addTemplate("array_unmarshal", `
// for {{.Index}} := range {{.Name}} {
//    {{.InnerCode}}

//    if err != nil {
//       return err
//    }
// }`)

// 	// Slice
// 	addTemplate("slice_marshal", `
// {
// buf.WriteUint64(uint64(len({{.Name}})))
// for {{.Index}} := range {{.Name}} {
//    {{.InnerCode}}
// }
// }`)

// 	addTemplate("slice_unmarshal", `
// {
// length, err := buf.ReadUint64()
// if err != nil { return err }
// for {{.Index}} := 0; {{.Index}} < int(length); {{.Index}}++ {
//    var {{.VarName}} {{.Type}}
//    {{.InnerCode}}
//    if err != nil {
//       return err
//    }

//    {{.Name}} = append({{.Name}}, {{.VarName}})
// }
// }`)


// 	// Map
// 	addTemplate("map_marshal", `
// {
// buf.WriteUint64(uint64(len({{.Name}})))

// for {{.KeyIdx}}, {{.ValIdx}} := range {{.Name}} {
//    {{.InnerCode}}
// }

// }`)

// 	addTemplate("map_unmarshal", `
// {
// length, err := buf.ReadUint64()
// if err != nil { return err }

// if {{.Name}} == nil {
// {{.Name}} = make({{.Type}})
// }

// for {{.Index}} := 0; {{.Index}} < int(length); {{.Index}}++ {
//    var {{.KeyVar}} {{.KeyType}}
//    var {{.ValVar}} {{.ValType}}

//    {{.InnerCode}}
//    if err != nil {
//       return err
//    }

//    {{.Name}}[{{.KeyVar}}] = {{.ValVar}}
// }
// }`)


// 	// Alias
// 	addTemplate("alias_marshal", `
// {
//    {{.ValName}} := {{.Type}}({{.Name}})
//    {{.InnerCode}}

// }`)

// 	addTemplate("alias_unmarshal", `
// {
//    var {{.ValName}} {{.Type}}
//    {{.InnerCode}}
//    *{{.Name}} = {{.AliasType}}({{.ValName}})
// }`)

// 	// Union
// 	addTemplate("union_marshal", `
//    // codUnion := cod.Union(t)
//    // rawVal := codUnion.GetRawValue()
//    rawVal := t.Get()
//    if rawVal == nil {
//       buf.WriteUint8(0) // Zero tag indicates nil
//       return
//    }

//    switch sv := rawVal.(type) {
//    {{.InnerCode}}
//    default:
//       panic("unknown type placed in union")
//    }
// `)

// 	addTemplate("union_unmarshal", `
//    // codUnion := cod.Union(*t)

//    tagVal := buf.ReadUint8()

//    switch tagVal {
//    case 0: // Zero tag indicates nil
//       return nil

//    {{.InnerCode}}
//    default:
//       panic("unknown type placed in union")
//    }
//    return err
// `)

// 	// Union cases
// 	addTemplate("union_case_marshal", `
//    case {{.Type}}:
//       buf.WriteUint8({{.Tag}})
//       sv.EncodeCod(buf)
//       // {{.InnerCode}}
// `)

// 	addTemplate("union_case_unmarshal", `
//    case {{.Tag}}:
//       var decoded {{.Type}}
//       err = decoded.DecodeCod(buf)
//       if err != nil { return err }
//       // codUnion.PutRawValue(decoded)
//       t.Set(decoded)
// `)

// 	// Union getters and setters
// 	addTemplate("union_getter", `
// func (t {{.Name}}) Get() any {
//    codUnion := cod.Union(t)
//    rawVal := codUnion.GetRawValue()
//    return rawVal

//    // switch rawVal.(type) {
//    // {{.InnerCode}}
//    // default:
//    //    panic("unknown type placed in union")
//    // }
// }
// `)

// 	addTemplate("union_constructor", `
// func New{{.Name}}(v any) {{.Name}} {
//    var ret {{.Name}}
//    ret.Set(v)
//    return ret
// }
// `)

// // TODO: You could theoretically check and panic if the user passes in an incorrect value
// 	addTemplate("union_setter", `
// func (t *{{.Name}}) Set(v any) {
//    codUnion := cod.Union(*t)
//    codUnion.PutRawValue(v)
//    *t = {{.Name}}(codUnion)

//    // switch tagVal {
//    // case 0: // Zero tag indicates nil
//    //    return nil

//    // {{.InnerCode}}
//    // default:
//    //    panic("unknown type placed in union")
//    // }
//    // return err
// }
// `)

}
