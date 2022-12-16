package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/go-openapi/loads"
	"go.uber.org/zap"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	v := config()
	logger := newLogger(v)
	validateConfig(v, logger)

	neofsAPI, err := newNeofsAPI(ctx, logger, v)
	if err != nil {
		logger.Fatal("init frostfs", zap.Error(err))
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		logger.Fatal("init spec", zap.Error(err))
	}

	serverCfg := serverConfig(v)
	serverCfg.SuccessfulStartCallback = neofsAPI.StartCallback

	api := operations.NewFrostfsRestGwAPI(swaggerSpec)
	server := restapi.NewServer(api, serverCfg)
	defer func() {
		if err = server.Shutdown(); err != nil {
			logger.Error("shutdown", zap.Error(err))
		}
	}()

	server.ConfigureAPI(neofsAPI.Configure)
	neofsAPI.RunServices()

	// serve API
	if err = server.Serve(); err != nil {
		logger.Fatal("serve", zap.Error(err))
	}
}
