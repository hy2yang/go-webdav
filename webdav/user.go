package webdav

import (
	"regexp"
	"strings"
)

// Rule is a dissalow/allow rule.
type Rule struct {
	Regex  bool
	Allow  bool
	Path   string
	Regexp *regexp.Regexp
}

// User contains the settings of each user.
type User struct {
	Username string
	Password string
	Scope    string
	Modify   bool
	Rules    []*Rule
}

// Allowed checks if the user has permission to access a directory/file
func (u User) Allowed(url string) bool {

	for i := 0; i <= len(u.Rules)-1; i++ {
		if u.Rules[i].match(url) {
			return u.Rules[i].Allow
		}
	}

	return true
}

func (r Rule) match(url string) bool {
	return (r.Regex && r.Regexp.MatchString(url)) || (!r.Regex && strings.HasPrefix(url, r.Path))
}
