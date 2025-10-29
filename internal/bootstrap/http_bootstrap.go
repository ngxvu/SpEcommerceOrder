package bootstrap

import (
	"fmt"
	"log"
	"net/http"
	"order/pkg/core/configloader"
)

func StartServer(router http.Handler, config *configloader.Config) {

	serverPort := fmt.Sprintf(":%s", config.ServerPort)
	s := &http.Server{
		Addr:    serverPort,
		Handler: router,
	}
	log.Println("Server started on port", serverPort)
	if err := s.ListenAndServe(); err != nil {
		_ = fmt.Errorf("failed to start server on port %s: %w", serverPort, err)
		panic(err)
	}
}
