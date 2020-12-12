package cmd

import (
	"log"
	"net"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/hy2yang/go-webdav/webdav"
)

func init() {
	var cfgFileName string

	cobra.OnInitialize(initConfig(&cfgFileName))

	flags := rootCmd.Flags()
	flags.StringVarP(&cfgFileName, "config", "c", "", "config file path")
	flags.BoolP("tls", "t", false, "enable tls")
	flags.Bool("auth", true, "enable auth")
	flags.String("cert", "cert.pem", "TLS certificate")
	flags.String("key", "key.pem", "TLS key")
	flags.StringP("address", "a", "0.0.0.0", "address to listen to")
	flags.StringP("port", "p", "8080", "port to listen to")

}

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "webdav",
	Short: "start WebDAV server",
	Long: `Will look for a configuration file called
				config.{json, toml, yaml, yml} in the following directories:

				- ./

				if flag "config" is not set.
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
		handler := webdav.HandlerFromConfig(parseConfig(flags))

		// Builds the address and a listener.
		socket := getValAsString(flags, "address") + ":" + getValAsString(flags, "port")
		listener, err := net.Listen("tcp", socket)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Webdav server istening on", listener.Addr().String())

		// Starts the server.
		if getValAsBool(flags, "tls") {
			if err := http.ServeTLS(listener, handler, getValAsString(flags, "cert"), getValAsString(flags, "key")); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := http.Serve(listener, handler); err != nil {
				log.Fatal(err)
			}
		}
	},
}
