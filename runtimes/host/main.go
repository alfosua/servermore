package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"servermore/host/guests"
	"servermore/host/options"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := ServeConcurrent(host)
	defer ShutdownServer(server)

	cmd := setupWorkerProcess(options)
	defer interruptProcess(cmd)

	<-ctx.Done()
	stop()

	return 0
}

func setupWorkerProcess(options options.HostOptions) *exec.Cmd {
	cmd, err := guests.StartWorker(options)

	if err != nil {
		panic(err)
	}
	return cmd
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
		panic(err)
	}

}
