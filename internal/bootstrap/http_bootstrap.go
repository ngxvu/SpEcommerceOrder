package bootstrap

import (
	"fmt"
	"log"
	"net/http"
	"order_service/pkg/core/configloader"
)

func StartServer(router http.Handler, cfg *configloader.Config) (*http.Server, <-chan error) {

	port := fmt.Sprintf(":%s", cfg.ServerPort)
	s := &http.Server{Addr: port, Handler: router}
	errCh := make(chan error, 1)

	go func() {
		log.Printf("HTTP server starting on %s", port)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	return s, errCh
}
