package main

import (
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/Daniel-EXLTK/exltk-rpc/internal/services/r0d0"
)

func main() {
	log.Println("🚀 EXLTK-RPC: r0d0-service starting...")

	// Create and register the R0D0 service
	service := r0d0.NewR0D0Service()

	// Register the service with RPC
	err := rpc.Register(service)
	if err != nil {
		log.Fatal("Failed to register R0D0 service:", err)
	}

	// Create HTTP handler for JSON-RPC
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create JSON-RPC codec
		codec := jsonrpc.NewServerCodec(&httpConn{r: r, w: w})
		defer codec.Close()

		// Serve the RPC request
		rpc.ServeCodec(codec)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "r0d0-service", "version": "1.0.0"}`))
	})

	log.Println("R0D0 Service listening on :8501")
	log.Println("Available methods:")
	log.Println("  - R0D0Service.Describe")
	log.Println("  - R0D0Service.DiscoveryStart")
	log.Println("  - R0D0Service.DiscoveryContinue")
	log.Println("  - R0D0Service.DiscoveryComplete")
	log.Println("Health check: http://localhost:8501/health")

	log.Fatal(http.ListenAndServe(":8501", nil))
}

// httpConn adapts HTTP request/response to io.ReadWriteCloser for JSON-RPC
type httpConn struct {
	r *http.Request
	w http.ResponseWriter
}

func (c *httpConn) Read(p []byte) (n int, err error) {
	return c.r.Body.Read(p)
}

func (c *httpConn) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *httpConn) Close() error {
	return c.r.Body.Close()
}
