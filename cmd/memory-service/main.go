package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Daniel-EXLTK/exltk-rpc/internal/memory"
)

func main() {
	ctx := context.Background()
	firestore, err := memory.NewFirestoreMemory(ctx)
	if err != nil {
		log.Fatalf("Error inicializando Firestore: %v", err)
	}
	api := memory.NewMemoryAPI(firestore)

	http.HandleFunc("/memory/save", api.SaveHandler)
	http.HandleFunc("/memory/get", api.GetHandler)
	http.HandleFunc("/memory/delete", api.DeleteHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("🚀 Memory Service escuchando en http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error en servidor HTTP: %v", err)
	}
}
