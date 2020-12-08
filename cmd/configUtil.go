package cmd

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hy2yang/go-webdav/webdav"

	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
)

func initConfig() {
	if cfgFile == "" {
		v.AddConfigPath(".")
		v.SetConfigName("config")
	} else {
		v.SetConfigFile(cfgFile)
	}

	v.SetEnvPrefix("WD")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(v.ConfigParseError); ok {
			panic(err)
		}
		cfgFile = "No config file used"
	} else {
		cfgFile = "Using config file: " + v.ConfigFileUsed()
	}
}

func readConfig(flags *pflag.FlagSet) *webdav.Config {
	cfg := &webdav.Config{
		Auth: getOptB(flags, "auth"),
		Cors: webdav.CorsCfg{
			Enabled:     false,
			Credentials: false,
		},
		Users: map[string]*webdav.User{},
	}

	rawUsers := v.Get("users")
	if users, ok := rawUsers.([]interface{}); ok {
		parseUsers(users, cfg)
	}

	rawCors := v.Get("cors")
	if cors, ok := rawCors.(map[string]interface{}); ok {
		parseCors(cors, cfg)
	}

	if len(cfg.Users) != 0 && !cfg.Auth {
		log.Print("Users will be ignored due to auth=false")
	}

	return cfg
}

func parseRules(raw []interface{}) []*webdav.Rule {
	rules := []*webdav.Rule{}

	for _, v := range raw {
		if r, ok := v.(map[interface{}]interface{}); ok {
			rule := &webdav.Rule{
				Regex: false,
				Allow: false,
				Path:  "",
			}

			if regex, ok := r["regex"].(bool); ok {
				rule.Regex = regex
			}

			if allow, ok := r["allow"].(bool); ok {
				rule.Allow = allow
			}

			path, ok := r["path"].(string)
			if !ok {
				continue
			}

			if rule.Regex {
				rule.Regexp = regexp.MustCompile(path)
			} else {
				rule.Path = path
			}

			rules = append(rules, rule)
		}
	}

	return rules
}

func loadFromEnv(v string) (string, error) {
	v = strings.TrimPrefix(v, "{env}")
	if v == "" {
		return "", errors.New("no environment variable specified")
	}

	v = os.Getenv(v)
	if v == "" {
		return "", errors.New("the environment variable is empty")
	}

	return v, nil
}

func parseUsers(raw []interface{}, c *webdav.Config) {
	var err error
	for _, entry := range raw {
		if u, ok := entry.(map[interface{}]interface{}); ok {

			username, ok := u["username"].(string)
			if !ok {
				log.Fatal("user needs an username")
			}
			if strings.HasPrefix(username, "{env}") {
				username, err = loadFromEnv(username)
				checkErr(err)
			}

			password, ok := u["password"].(string)
			if !ok {
				if numPwd, ok := u["password"].(int); ok {
					password = strconv.Itoa(numPwd)
				} else {
					password = ""
				}
			}
			if strings.HasPrefix(password, "{env}") {
				password, err = loadFromEnv(password)
				checkErr(err)
			}

			rules, ok := u["rules"]
			parsedRules := []*webdav.Rule{}
			if ok {
				if rawRules, ok := rules.([]interface{}); ok {
					parsedRules = parseRules(rawRules)
				} else {
					log.Fatal("error parsing rules of user: ", username)
				}
				// rules are not required, but fails if rules exist and got errors when parsing
			}

			user := &webdav.User{
				Username: username,
				Password: password,
				Scope:    u["scope"].(string),
				Modify:   u["modify"].(bool),
				Rules:    parsedRules,
			}

			c.Users[username] = user

		}
	}
}

func parseCors(cfg map[string]interface{}, c *webdav.Config) {
	cors := webdav.CorsCfg{
		Enabled:     cfg["enabled"].(bool),
		Credentials: cfg["credentials"].(bool),
	}

	cors.AllowedHeaders = corsProperty("allowed_headers", cfg)
	cors.AllowedHosts = corsProperty("allowed_hosts", cfg)
	cors.AllowedMethods = corsProperty("allowed_methods", cfg)
	cors.ExposedHeaders = corsProperty("exposed_headers", cfg)

	c.Cors = cors
}

func corsProperty(property string, cfg map[string]interface{}) []string {
	var def []string

	if property == "exposed_headers" {
		def = []string{}
	} else {
		def = []string{"*"}
	}

	if allowed, ok := cfg[property].([]interface{}); ok {
		items := make([]string, len(allowed))

		for idx, a := range allowed {
			items[idx] = a.(string)
		}

		if len(items) == 0 {
			return def
		}
		return items
	}

	return def
}
