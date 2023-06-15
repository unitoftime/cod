Cod is a schemaless serialization library for generating serialization code for your structs. This is for generating code that encodes and decodes data as quickly as possible. By default we use variable length encoding for all unsigned and signed integers larger than 16 bits.

### Install
You just need to run this to get the binary: `go install github.com/unitoftime/cod/cmd/cod`
You will also need to add ~/go/bin/ (or windows equivalent is) to your path so you can reference binaries from there

You can then add `//go:generate cod` to one of your go files in your package. This will run the cod binary every time you execute `go generate`. Finally you can tag structures that you want to generate code for as follows:
1. Structs: `//cod:struct`
2. Unions: `//cod:union <CSV List of Unionable Types>`

#### Disclaimers
1. AST Parsing and code generation is tricky to get right. If you do find a situation where the code is not generated correctly, please let me know by opening an issue.
2. Map serialization is not deterministic. This is because looping over a map is not deterministic. I can maybe add this in the future if people want it.
3. There's no versioning info included in the serialized data by default. If you want to support multiple encodings, then you'll need to include them all in a tagged union
4. Currently, tagged unions can support a maximum of 255 different types

### Supports
1. Basic data types
2. Custom Structs
3. Slices and Maps
4. "Hand-Crafted" Encoders and Decoders (Just implement the two normally generated functions)
5. Serializes private fields by default (TODO to be able to turn that off)

### TODOs
1. Multiple backends (ie different swappable serialization schemes)
2. Generated Size functions (to let you measure the size of the thing you want to serialize)
3. Ability to selectively turn off variable length encoding with a tag: `cod:"fixed"` (or something)
4. Ability to prevent a field from serializing (ie disable fields)
5. Automatic bitpacking for structs that are just a long list of bools (or similar)

### Syntax
#### Custom Types
Add the `//cod:struct` to indicate that its a struct, or create a serializable union with `//cod:union <CSV list of union types>`:

```
//cod:struct
type Person struct {
    Name string
    Age uint8
    FavoriteThings map[string]Thing
    SomeListOfNumbers []int64
}

//cod:union Ball, Hat
type Thing cod.Union

//cod:struct
type Ball struct {
    Color string
}

//cod:struct
type Hat struct {
    Material string
}
```

#### Generated Functions
All types will have two methods generated for them:
1. `EncodeCod([]byte) []byte`
2. `DecodeCod([]byte) (int, error)`

Unions will also get the following methods
1. `Get() any // Guaranteed to return one of the unionable types or nil`
2. `Set(any)  // Must pass in only a unionable type or nil`
3. `New<TYPE>(v any) <TYPE> // A Constructor: Where <TYPE> is the name of the union`


```
// ---------------
// For cod:structs
// ---------------

func (t Person) EncodeCod(bs []byte) []byte {
    // ... Generated Code ...
}
func (t *Person) DecodeCod(bs []byte) (int, error) {
    // ... Generated Code ...
}

// ... Other Structs ...

// ---------------
// For cod:unions
// ---------------

func (t <TYPE>) EncodeCod(bs []byte) []byte {
    // ... Generated Code ...
}

func (t *<TYPE>) DecodeCod(bs []byte) (int, error) {
    // ... Generated Code ...
}

func (t <TYPE>) Get() any {
    // ... Generated Code ...
}

func (t *<TYPE>) Set(v any) {
    // ... Generated Code ...
}

func New<TYPE>(v any) <TYPE> {
    // ... Generated Code ...
}
```

### Inspirations
1. https://github.com/mus-format/mus-go
2. https://github.com/alecthomas/go_serialization_benchmarks
3. https://github.com/Kelindar/binary
