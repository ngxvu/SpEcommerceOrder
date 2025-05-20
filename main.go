package main

import (
	"context"
	"emission/conf"
	"emission/internal/route"
	"emission/internal/utils"
	"emission/pkg/http/logger"
)

const (
	APPNAME = "emission"
)

func main() {
	conf.SetEnv()
	logger.Init(APPNAME)
	utils.LoadMessageError()
	app := route.NewService()
	err := app.Start(context.Background())
	if err != nil {
		logger.DefaultLogger.Error(err)
	}
}
