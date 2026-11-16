package input

import "strings"

// Sanitize rimuove o sostituisce un insieme fisso di caratteri "pericolosi".
//
// ATTENZIONE: non è una protezione da XSS né da SQL injection. Contro l'XSS
// usa l'escaping in output (html/template); contro la SQL injection usa le
// query parametrizzate (i builder di orm). Il parametro level è attualmente
// ignorato ed è mantenuto per compatibilità.
func Sanitize(str string, level int) string {
	// Array di caratteri da rimuovere
	arr := []string{`"`, `\`, "@", "%", "_", "^", "*", "(", ")", "[", "]", "{", "}", "|", "/", "$", "#", "<", "&", ">", "+", "--"}

	// Rimuovi i caratteri specificati dall'array
	for _, char := range arr {
		str = strings.ReplaceAll(str, char, "")
	}

	// Sostituisci singolo apice con apice tipografico
	str = strings.ReplaceAll(str, "'", "’")

	// Sostituisci punto e virgola con una virgola
	str = strings.ReplaceAll(str, ";", ",")

	return str
}
