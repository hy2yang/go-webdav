package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
)

// precedence: flags -> viper (env, config) -> default
func getValAsString(flags *pflag.FlagSet, key string) string {
	value, _ := flags.GetString(key)

	// If set on Flags, use it.
	if flags.Changed(key) {
		return value
	}

	// If set through viper (env, config), return it.
	if v.IsSet(key) {
		return v.GetString(key)
	}

	// Otherwise use default value on flags.
	return value
}

// precedence: flags -> viper (env, config) -> default
func getValAsBool(flags *pflag.FlagSet, key string) bool {
	value, _ := flags.GetBool(key)

	// If set on Flags, use it.
	if flags.Changed(key) {
		return value
	}

	// If set through viper (env, config), return it.
	if v.IsSet(key) {
		return v.GetBool(key)
	}

	// Otherwise use default value on flags.
	return value
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
