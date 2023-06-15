package cod

type Union struct {
	value any
}
func (u Union) GetRawValue() any {
	return u.value
}
func (u *Union) PutRawValue(v any) {
	u.value = v
}
