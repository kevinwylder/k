package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kevinwylder/k/api"
	"github.com/kevinwylder/k/fs"
)

func main() {
	data, err := fs.NewStorageDir("server", ".", "")
	if err != nil {
		log.Fatalf("storage dir: %v", err)
	}

	server := api.NewServer(data)

	http.ListenAndServe(os.Getenv("PORT"), server)
}
