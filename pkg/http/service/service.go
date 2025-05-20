package service

import (
	"context"
	"emission/pkg/http/db"
	ginext2 "emission/pkg/http/ginext"
	"emission/pkg/http/logger"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/caarlos0/env/v6"
)

type BaseApp struct {
	Config     *AppConfig
	Name       string
	Version    string
	Router     *gin.Engine
	HttpServer *http.Server

	listener       net.Listener
	initialized    bool
	healthDisabled bool
}

func NewApp(name, version string) *BaseApp {
	app := &BaseApp{
		Name:           name,
		Version:        version,
		Router:         gin.New(),
		HttpServer:     &http.Server{},
		Config:         NewAppConfig(),
		healthDisabled: false,
	}

	app.HttpServer.Handler = app.Router

	return app
}

func (app *BaseApp) DisableHealthEndpoint() {
	app.healthDisabled = true
}

func (app *BaseApp) Initialize() error {
	if err := env.Parse(app); err != nil {
		return err
	}

	app.HttpServer.ReadTimeout = time.Duration(app.Config.ReadTimeout) * time.Second

	// wait to gin 1.8 to support this
	//_ = app.Router.SetTrustedProxies(app.Config.TrustedProxy)
	app.Router.ForwardedByClientIP = true

	// register default middlewares
	app.Router.Use(
		ginext2.RequestIDMiddleware,
		ginext2.AccessLogMiddleware(app.Config.Env),
		ginext2.CreateErrorHandler(app.Config.Debug),
	)

	// register routes
	if !app.healthDisabled {
		healthHandler := app.HealthHandler()
		app.Router.GET("/status", healthHandler)
		app.Router.GET("/status-q", healthHandler)
	}

	app.Router.NoRoute(ginext2.NotFoundHandler)

	if app.Config.EnableDB {
		err := db.OpenDefault(app.Config.DB)
		if err != nil {
			return errors.New("failed to open default DB: " + err.Error())
		}
	}

	app.initialized = true

	return nil
}

// HealthHandler makes health check handler
func (app *BaseApp) HealthHandler() gin.HandlerFunc {
	rsp := struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{
		Name:    app.Name,
		Version: app.Version,
	}

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, rsp)
	}
}

func (app *BaseApp) StartTLS(ctx context.Context, certPath string, keyPath string) error {
	l := logger.Tag("BaseApp.Start")
	var err error

	if !app.initialized {
		if err = app.Initialize(); err != nil {
			return errors.New("failed to initialize app: " + err.Error())
		}
	}

	if app.listener, err = net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%d", app.Config.Port)); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	errCh := make(chan error, 1)

	go func() {
		l.Printf("start listening on %s", app.listener.Addr().String())
		if err := app.HttpServer.ServeTLS(app.listener, certPath, keyPath); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}

		// no error, close channel
		close(errCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		defer func() {
			l.Info("shutting down http server ...")
			shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = app.HttpServer.Shutdown(shutCtx)
			cancel()
		}()
		select {
		case gotSignal, ok := <-signalCh:
			if !ok {
				// channel close
				return
			}
			l.Printf("got signal: %v", gotSignal)
			return
		case <-ctx.Done():
			l.Printf("context has done")
			return
		}
	}()

	go func() {
		l.Printf("start listening debug server on port %d", app.Config.DebugPort)
		_ = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(app.Config.DebugPort), nil)
	}()

	return <-errCh
}

func (app *BaseApp) Start(ctx context.Context) error {
	l := logger.Tag("BaseApp.Start")
	var err error

	if !app.initialized {
		if err = app.Initialize(); err != nil {
			return errors.New("failed to initialize app: " + err.Error())
		}
	}

	if app.listener, err = net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%d", app.Config.Port)); err != nil {
		return errors.New("failed to listen: " + err.Error())
	}

	errCh := make(chan error, 1)

	go func() {
		l.Printf("start listening on %s", app.listener.Addr().String())
		if err := app.HttpServer.Serve(app.listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}

		// no error, close channel
		close(errCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		defer func() {
			l.Info("shutting down http server ...")
			shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = app.HttpServer.Shutdown(shutCtx)
			cancel()
		}()

		select {
		case gotSignal, ok := <-signalCh:
			if !ok {
				// channel close
				return
			}
			l.Printf("got signal: %v", gotSignal)
			return
		case <-ctx.Done():
			l.Printf("context has done")
			return
		}
	}()

	go func() {
		l.Printf("start listening debug server on port %d", app.Config.DebugPort)
		_ = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(app.Config.DebugPort), nil)
	}()

	return <-errCh
}

func (app *BaseApp) Listener() net.Listener {
	return app.listener
}

func (app *BaseApp) GetDB() *gorm.DB {
	if !app.initialized {
		err := app.Initialize()
		if err != nil {
			panic(err)
		}
	}
	return db.GetDB()
}
