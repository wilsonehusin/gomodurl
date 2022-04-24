package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"go.husin.dev/gomodurl"
)

func main() {
	gomodurl.EnableLogger(os.Stderr)

	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "error: %s", err.Error())
		os.Exit(1)
	}
}

const (
	configEnvVar = "GOMODURL_CONFIG"
)

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	configPath := os.Getenv(configEnvVar)
	h, err := gomodurl.HTTPHandler(ctx, configPath)
	if err != nil {
		return fmt.Errorf("generating handler: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	s := &http.Server{
		Addr: ":" + port,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: h,
	}

	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()
}
