package input

import (
	"strings"
)

// ContainsXSS segnala la presenza di alcuni pattern XSS comuni.
//
// ATTENZIONE: è una blacklist e si aggira facilmente (es. <details ontoggle=...>,
// codifiche HTML/URL, attributi on* non elencati). Usala solo come controllo
// euristico in più: la vera difesa è l'escaping in output con html/template
// (ed eventualmente una Content-Security-Policy).
//
// Deprecated: non affidarti a una blacklist. Valida l'input secondo il suo
// formato (pacchetto validator) ed esegui l'escaping in output con
// html/template, o html.EscapeString fuori dai template.
func ContainsXSS(input string) bool {
	lowerInput := strings.ToLower(input)

	// Lista di pattern comuni usati negli attacchi XSS
	xssPatterns := []string{
		"<script",
		"</script",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"<iframe",
		"</iframe",
		"<img",
		"<svg",
		"<object",
		"<embed",
		"<link",
		"<style",
		"document.cookie",
		"document.write",
		"window.location",
		"eval(",
		"settimeout(",
		"setinterval(",
		"alert(",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	return false
}
