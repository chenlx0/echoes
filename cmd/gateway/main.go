package main

import (
	"flag"
	"os"

	"github.com/chenlx0/echoes/internal/config"
	"github.com/chenlx0/echoes/internal/lg"

	"github.com/chenlx0/echoes/internal/httpserver"
)

var configFile *string

func init() {
	configFile = flag.String("config", "resources/echoes.yaml", "specify yaml config file")
	flag.Parse()
}

func main() {
	conf, err := config.ReadConfig(*configFile)
	if err != nil {
		lg.LogFatal("Read config: ", err)
		os.Exit(1)
	}
	httpserver.Run(conf)
}
