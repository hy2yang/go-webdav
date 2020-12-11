package cmd

import (
	"errors"
	"log"
	"os"
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

func parseConfig(flags *pflag.FlagSet) *webdav.Config {
	cfg := &webdav.Config{
		Auth: getValAsBool(flags, "auth"),
		Cors: webdav.CorsCfg{
			Enabled:     false,
			Credentials: false,
		},
		Users: parseUsers(), //parsed users
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

func parseUsers() map[string]*webdav.User {
	res := map[string]*webdav.User{}

	var users []webdav.User
	err := v.UnmarshalKey("users", &users)

	if err != nil {
		log.Fatal("error parsing user configs")
	} else {
		for _, user := range users {
			res[user.Username] = &user
		}
	}

	return res
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
