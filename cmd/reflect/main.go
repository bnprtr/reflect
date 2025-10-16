package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bnprtr/reflect/internal/config"
	"github.com/bnprtr/reflect/internal/descriptor"
	"github.com/bnprtr/reflect/internal/server"
	"github.com/bnprtr/reflect/internal/server/theme"
	"github.com/bnprtr/reflect/internal/watcher"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	protoRoot := flag.String("proto-root", "", "root directory containing .proto files")
	themeName := flag.String("theme", "default", "theme name (default, minimal, high-contrast, ocean, forest, sunset, monochrome)")
	themeFile := flag.String("theme-file", "", "path to custom theme file (JSON or YAML)")
	configPath := flag.String("config", "", "path to reflect.yaml configuration file (optional)")
	var protoIncludes []string
	flag.Func("proto-include", "include path for proto imports (can be specified multiple times)", func(value string) error {
		protoIncludes = append(protoIncludes, value)
		return nil
	})
	devMode := flag.Bool("dev", false, "enable development mode with hot reloading")
	flag.Parse()

	ctx := context.Background()

	// Load configuration if specified
	var cfg *config.Config
	if *configPath != "" {
		var err error
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config from %q: %v", *configPath, err)
		}
		log.Printf("Loaded configuration from %q with %d environment(s)", *configPath, len(cfg.Environments))
	}

	// Load protobuf descriptors if proto-root is specified
	var reg *descriptor.Registry
	if *protoRoot != "" {
		var err error
		reg, err = descriptor.LoadDirectory(ctx, *protoRoot, protoIncludes)
		if err != nil {
			log.Fatalf("Failed to load proto files from %q: %v", *protoRoot, err)
		}
		log.Printf("Loaded proto files from %q", *protoRoot)
	}

	// Load theme
	var selectedTheme *theme.Theme
	var err error

	if *themeFile != "" {
		// Load theme from file
		selectedTheme, err = theme.LoadThemeFromFile(*themeFile)
		if err != nil {
			log.Fatalf("Failed to load theme from file %q: %v", *themeFile, err)
		}
		log.Printf("Loaded theme %q from file: %s", selectedTheme.Name, *themeFile)
	} else {
		// Load built-in theme
		selectedTheme = theme.GetThemeByName(*themeName)
		log.Printf("Using theme: %s", selectedTheme.Name)
	}

	srv, err := server.NewWithTheme(reg, selectedTheme, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Setup hot reloading if in dev mode and proto-root is specified
	if *devMode && *protoRoot != "" {
		log.Println("Dev mode enabled - watching for proto file changes")

		// Create context for watcher
		watcherCtx, cancelWatcher := context.WithCancel(ctx)
		defer cancelWatcher()

		// Create watcher with reload function
		w, err := watcher.New(*protoRoot, func() {
			// Reload proto files
			newReg, err := descriptor.LoadDirectory(ctx, *protoRoot, protoIncludes)
			if err != nil {
				log.Printf("Failed to reload proto files: %v", err)
				return
			}
			// Update server with new registry
			srv.SetRegistry(newReg)
		})
		if err != nil {
			log.Fatalf("Failed to create file watcher: %v", err)
		}
		defer w.Close()

		// Start watcher in background
		go w.Start(watcherCtx)
	}

	// Setup graceful shutdown
	httpServer := &http.Server{
		Addr:    *addr,
		Handler: srv,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("listening on %s", *addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Shutdown with timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
