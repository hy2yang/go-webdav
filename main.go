package main

import (
	"runtime"

	"github.com/hy2yang/go-webdav/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
