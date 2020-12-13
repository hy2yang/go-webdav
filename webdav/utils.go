package webdav

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	modMethods = map[string]struct{}{
		"PUT":    {},
		"POST":   {},
		"MKCOL":  {},
		"DELETE": {},
		"COPY":   {},
		"MOVE":   {},
	}
	pwCache = map[string]string{} // valid saved-input value pairs
)

func checkPassword(saved, input string) bool {
	if v, pr := pwCache[saved]; pr {
		return v == input
	}

	var res bool
	if strings.HasPrefix(saved, "{bcrypt}") {
		res = bcrypt.CompareHashAndPassword(
			[]byte(strings.TrimPrefix(saved, "{bcrypt}")), []byte(input)) == nil
	} else {
		res = (saved == input)
	}
	if res {
		pwCache[saved] = input
	}

	return res
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
	if !u.Allowed(r.URL.Path) || (!u.Modify && isModMethod(r.Method)) {
		return false
	}
	return true
}

func isModMethod(method string) bool {
	_, ok := modMethods[method]
	return ok
}
