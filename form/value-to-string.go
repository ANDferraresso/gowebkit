package form

import "github.com/ANDferraresso/gowebkit/cast"

// ValToStr converte un valore in stringa.
//
// Deprecated: usa cast.ValToStr; questa funzione resta per compatibilità.
func ValToStr(v interface{}) string {
	return cast.ValToStr(v)
}
