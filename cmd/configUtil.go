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

func initConfig(cfgFileName *string) func() {
	return func() {
		if *cfgFileName != "" {
			v.SetConfigFile(*cfgFileName)
		} else {
			v.AddConfigPath(".")
			v.SetConfigName("config")
		}

		v.SetEnvPrefix("WD")
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
	}
}

func parseConfig(flags *pflag.FlagSet) *webdav.Config {
	cfg := &webdav.Config{
		Auth:  getValAsBool(flags, "auth"),
		Cors:  parseCors(),
		Users: parseUsers(),
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

func parseCors() webdav.CorsCfg {
	v.SetDefault("cors.enabled", false)
	v.SetDefault("cors.credentials", false)
	v.SetDefault("cors.allowed_headers", "*")
	v.SetDefault("cors.allowed_hosts", "*")
	v.SetDefault("cors.allowed_methods", "*")
	v.SetDefault("cors.exposed_headers", "")

	var res webdav.CorsCfg

	err := v.UnmarshalKey("cors", &res)

	if err != nil {
		log.Fatal("error parsing user configs")
	}

	return res
}

func loadFromEnv(v string) (string, error) {
	var trimmed = strings.TrimPrefix(v, "{env}")
	if trimmed == "" {
		return "", errors.New("no env prefix found specified in " + v)
	}

	var res = os.Getenv(trimmed)
	if res == "" {
		return "", errors.New("env " + trimmed + " is empty")
	}

	return res, nil
}
