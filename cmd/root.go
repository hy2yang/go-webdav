package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/hy2yang/go-webdav/webdav"
)

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)

	flags := rootCmd.Flags()
	flags.StringVarP(&cfgFile, "config", "c", "", "config file path")
	flags.BoolP("tls", "t", false, "enable tls")
	flags.Bool("auth", true, "enable auth")
	flags.String("cert", "cert.pem", "TLS certificate")
	flags.String("key", "key.pem", "TLS key")
	flags.StringP("address", "a", "0.0.0.0", "address to listen to")
	flags.StringP("port", "p", "0", "port to listen to")
}

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "webdav",
	Short: "A simple to use WebDAV server",
	Long: `If you don't set "config", it will look for a configuration file called
config.{json, toml, yaml, yml} in the following directories:

- ./
- /etc/webdav/

The precedence of the configuration values are as follows:

- flags
- environment variables
- configuration file
- defaults

The environment variables are prefixed by "WD_" followed by the option
name in caps. So to set "cert" via an env variable, you should
set WD_CERT.`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		handler := webdav.HandlerFromConfig(readConfig(flags))

		// Builds the address and a listener.
		address := getOpt(flags, "address") + ":" + getOpt(flags, "port")
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		// Tell the user the port in which is listening.
		fmt.Println("Listening on", listener.Addr().String())

		// Starts the server.
		if getOptB(flags, "tls") {
			if err := http.ServeTLS(listener, handler, getOpt(flags, "cert"), getOpt(flags, "key")); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := http.Serve(listener, handler); err != nil {
				log.Fatal(err)
			}
		}
	},
}
