package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bnprtr/reflect/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	_ = flag.Bool("dev", false, "dev mode (reserved)")
	flag.Parse()

	h, err := server.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, h))
}
