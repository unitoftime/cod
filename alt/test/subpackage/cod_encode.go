package subpackage

import (
	"github.com/unitoftime/cod/backend"
)

func (t Vec) EncodeCod(bs []byte) []byte {

	bs = backend.WriteVarUint64(bs, (t.X))

	bs = backend.WriteVarUint64(bs, (t.Y))

	return bs
}

func (t *Vec) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var decoded uint64
		decoded, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.X = (decoded)
	}

	{
		var decoded uint64
		decoded, nOff, err = backend.ReadVarUint64(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff
		t.Y = (decoded)
	}

	return n, err
}
