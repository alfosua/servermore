package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"syscall"

	"servermore/host/guests"
	"servermore/host/options"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Functions []string `json:"functions"`
}

func main() { os.Exit(run()) }

func run() int {
	options, err := options.ParseArgs()

	if err != nil {
		panic(err)
	}

	host := NewServermoreHost(options)

	cmd, err := guests.StartWorker(options)

	if err != nil {
		panic(err)
	}

	defer interruptProcess(cmd)

	go serve(host)

	signalReceived := WaitSignals(syscall.SIGINT, syscall.SIGTERM)

	switch signalReceived {
	case syscall.SIGINT, syscall.SIGTERM:
		log.Println("Received signal, shutting down...")
	default:
		log.Println("Received unknown signal.")
		return 1
	}

	return 0
}

func serve(host *ServermoreHost) {
	router := gin.Default()

	router.GET("/:funcName", host.ServerlessFunctionGet)

	router.GET("/internals/options", host.InternalOptionsGet)
	router.POST("/internals/worker", host.WorkerPost)

	router.Run(":8080")
}

func ReverseProxyOf(endpoint string, writer http.ResponseWriter, req *http.Request) error {
	remote, err := url.Parse("http://localhost:3000")
	if err != nil {
		return errors.New("invalid url")
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(proxyReq *http.Request) {
		proxyReq.Header = req.Header
		proxyReq.Host = remote.Host
		proxyReq.URL.Scheme = remote.Scheme
		proxyReq.URL.Host = remote.Host
		proxyReq.URL.Path = endpoint
	}

	proxy.ServeHTTP(writer, req)

	return nil
}

func interruptProcess(cmd *exec.Cmd) {
	log.Println("Interrupting worker")

	err := cmd.Process.Signal(os.Interrupt)

	if err != nil {
		log.Fatal(err)
	}

}
