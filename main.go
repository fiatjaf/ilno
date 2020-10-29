package main

import (
	"github.com/fiatjaf/ilno/config"
	"github.com/fiatjaf/ilno/logger"
	"github.com/fiatjaf/ilno/server"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var s config.Config
	err := envconfig.Process("", &s)
	if err != nil {
		logger.Fatal("couldn't process envconfig: %s", err)
	}

	switch s.LogLevel {
	case "DEBUG":
		logger.EnableDebug()
	}

	server.Serve(s)
}
