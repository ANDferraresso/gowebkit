package input

import (
	"strings"
)

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
