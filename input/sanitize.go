package sanitize

import "strings"

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
