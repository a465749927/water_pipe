package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"water_pipe/config"
	"water_pipe/proxy"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	server, err := proxy.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating proxy server: %v\n", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting proxy server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Water Pipe proxy server started (Node ID: %s)\n", cfg.Node.ID)
	fmt.Printf("Listening on %s\n", cfg.Node.ListenAddress)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down proxy server...")
	if err := server.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping proxy server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Proxy server stopped")
}
