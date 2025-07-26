package validator

import (
	"net"
	"regexp"
)

func IsDomain(value string, pars *[]string) bool {
	re := regexp.MustCompile(`^((xn--)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	return re.Match([]byte(value))
}

func IsEmail(value string, pars *[]string) bool {
	// re := regexp.MustCompile("^[-0-9a-zA-Z.+_]+@[-0-9a-zA-Z.+_]+\\.[a-zA-Z]{2,}$")
	// re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	re := regexp.MustCompile(`^[-0-9a-zA-Z.+_]+@[-0-9a-zA-Z.+_]+\.[a-zA-Z]{2,}$`)
	return re.Match([]byte(value))
}

func IsIPV4(value string, pars *[]string) bool {
	// re := regexp.MustCompile("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$")
	re := regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)
	if re.Match([]byte(value)) {
		if net.ParseIP(value) != nil {
			return true
		}
	}
	return false
}

func IsURL(value string, pars *[]string) bool {
	// re := regexp.MustCompile("^(https?:\\/\\/)?([\\da-zA-Z\\.-]+)\\.([a-zA-Z\\.]{2,6})([\\/\\w\\.-]*)*\\/?$
	// re := regexp.MustCompile(`^(https?:\/\/)?([\da-zA-Z\.-]+)\.([a-zA-Z\.]{2,6})([\/\w\.-]*)*\/?$`)
	// Sopra da errore exp: invalid or unsupported Perl syntax: `(?!` ...
	re := regexp.MustCompile(`^(https?://)?([\da-zA-Z.-]+)\.([a-zA-Z]{2,6})(/[^\s]*)?$`)
	return re.Match([]byte(value))
}
