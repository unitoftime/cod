package cod

const (
	UnionEmpty uint8 = 0
)

type EncoderDecoder interface {
	EncodeCod([]byte) []byte
	// DecodeCod([]byte) (int, error) // Doesn't fit b/c its a pointer receiver
}

type Union struct {
	value EncoderDecoder
}
func (u Union) GetRawValue() EncoderDecoder {
	return u.value
}
func (u *Union) PutRawValue(v EncoderDecoder) {
	u.value = v
}
