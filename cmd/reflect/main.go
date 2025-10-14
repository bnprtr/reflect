package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/bnprtr/reflect/internal/descriptor"
	"github.com/bnprtr/reflect/internal/server"
	"github.com/bnprtr/reflect/internal/server/theme"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	protoRoot := flag.String("proto-root", "", "root directory containing .proto files")
	themeName := flag.String("theme", "default", "theme name (default, minimal, high-contrast)")
	var protoIncludes []string
	flag.Func("proto-include", "include path for proto imports (can be specified multiple times)", func(value string) error {
		protoIncludes = append(protoIncludes, value)
		return nil
	})
	_ = flag.Bool("dev", false, "dev mode (reserved)")
	flag.Parse()

	ctx := context.Background()

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
	selectedTheme := theme.GetThemeByName(*themeName)
	log.Printf("Using theme: %s", selectedTheme.Name)

	h, err := server.NewWithTheme(reg, selectedTheme)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, h))
}
