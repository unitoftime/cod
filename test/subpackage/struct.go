package subpackage

//go:generate go run ../../cmd/cod

//cod:struct
type Vec struct {
	X, Y uint64
}


// func MapsEqual[K, V any](m1, m2 map[K]V) bool {
// 	if len(m1) != len(m2) { return false }
// 	for k, v := range m1 {
// 		aa, ok := m2[k]
// 		if !ok { return false }
// 		if aa != k { return false }
// 	}

// 	return true
// }
