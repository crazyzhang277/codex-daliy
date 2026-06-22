package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"nvidialimiter/internal/config"
	"nvidialimiter/internal/proxy"
	"nvidialimiter/internal/service"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			if err := service.Install(); err != nil {
				log.Fatal(err)
			}
			return
		case "uninstall":
			if err := service.Uninstall(); err != nil {
				log.Fatal(err)
			}
			return
		case "install-run":
			// service mode entrypoint
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exePath)

	cfg, err := config.Load(filepath.Join(baseDir, "whitelist.toml"))
	if err != nil {
		log.Printf("config load warning: %v", err)
		cfg = config.Default()
	}

	upstream := os.Getenv("UPSTREAM_BASE_URL")
	if upstream == "" {
		upstream = "https://integrate.api.nvidia.com/v1"
	}

	state := proxy.NewState(cfg)
	handler := proxy.NewHandler(state, upstream)

	srv := &http.Server{
		Addr:    "127.0.0.1:57321",
		Handler: handler,
	}

	go func() {
		log.Printf("listening on %s, upstream=%s", srv.Addr, upstream)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
	fmt.Println("shutdown complete")
}
