package server

import (
	"context"
	"flag"
	database "go_project/internal/db"
	"go_project/internal/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setupServerRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.ViewHanlder)
	mux.HandleFunc("/new", handlers.NewHandler)
	mux.HandleFunc("/create", handlers.CreateHandler)
	mux.HandleFunc("/delete", handlers.DeleteHandler)

	return mux
}

func newServer(port *string) *http.Server {
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: setupServerRoutes(),
	}
	return server
}

func startServer(server *http.Server) {
	go func() {
		log.Printf("Server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()
}

func endServer(server *http.Server, timeout time.Duration) {
	sigterm := make(chan os.Signal, 1) // package "closer" as an alternative
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-sigterm

	log.Println("Shutdown signal received, exiting...")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	defer database.ExitDatabase()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")
}

func Run() {
	// Get port value from flag
	portFlag := flag.String("port", "8080", "Server port")
	flag.Parse()

	// Initialize server
	server := newServer(portFlag)

	// Start server
	startServer(server)

	database.ConnectDatabase()

	// Gracefull shutdown
	endServer(server, 5*time.Second)
}
