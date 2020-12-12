package webdav

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var modMethods = map[string]struct{}{
	"PUT":    {},
	"POST":   {},
	"MKCOL":  {},
	"DELETE": {},
	"COPY":   {},
	"MOVE":   {},
}

func checkPassword(saved, input string) bool {
	if strings.HasPrefix(saved, "{bcrypt}") {
		savedPassword := strings.TrimPrefix(saved, "{bcrypt}")
		return bcrypt.CompareHashAndPassword([]byte(savedPassword), []byte(input)) == nil
	}

	return saved == input
}

func isAllowedHost(allowedHosts []string, origin string) bool {
	for _, host := range allowedHosts {
		if host == origin {
			return true
		}
	}
	return false
}

func userHasPermission(u *User, r *http.Request) bool {
	// Checks
	// 1. user permissions relatively to this PATH.
	// 2. if this request modified the files and the user doesn't have permission
	if !u.Allowed(r.URL.Path) || (isModMethod(r.Method) && !u.Modify) {
		return false
	}
	return true
}

func isModMethod(method string) bool {
	_, ok := modMethods[method]
	return ok
}
