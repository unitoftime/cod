package subpackage

//go:generate go run ../../cmd/cod

//cod:struct
type Vec struct {
	X, Y uint64
}
