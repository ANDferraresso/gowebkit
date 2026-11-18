package validator

import (
	"net"
	"regexp"
)

// Regex dei validator internet, compilate una volta sola.
var (
	domainRe = regexp.MustCompile(`^((xn--)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	emailRe  = regexp.MustCompile(`^[-0-9a-zA-Z.+_]+@[-0-9a-zA-Z.+_]+\.[a-zA-Z]{2,}$`)
	ipv4Re   = regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)
	// La versione con (?!...) non è supportata da RE2 (il motore regexp di Go).
	urlRe = regexp.MustCompile(`^(https?://)?([\da-zA-Z.-]+)\.([a-zA-Z]{2,6})(/[^\s]*)?$`)
)

func IsDomain(value string, pars *[]string) bool {
	return domainRe.MatchString(value)
}

func IsEmail(value string, pars *[]string) bool {
	return emailRe.MatchString(value)
}

func IsIPV4(value string, pars *[]string) bool {
	return ipv4Re.MatchString(value) && net.ParseIP(value) != nil
}

func IsURL(value string, pars *[]string) bool {
	return urlRe.MatchString(value)
}
