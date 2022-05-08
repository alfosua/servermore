package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func createServer(host *ServermoreHost) *http.Server {

	router := gin.Default()

	router.GET("/:funcName", host.ServerlessFunctionGet)

	router.GET("/internals/options", host.InternalOptionsGet)
	router.POST("/internals/worker", host.WorkerPost)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return server
}

func ServeConcurrent(host *ServermoreHost) *http.Server {

	server := createServer(host)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("listen: %s", err)
		}
	}()

	return server

}

func ShutdownServer(server *http.Server) {
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}

	log.Println("Server stopped")
}
