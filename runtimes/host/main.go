package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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

	"github.com/gin-gonic/gin"
)

type AppDef struct {
	Functions []string
}

func main() { os.Exit(run()) }

func run() int {
	options, err := options.ParseArgs()

	if err != nil {
		panic(err)
	}

	cmd, err := guests.StartWorker(options)

	if err != nil {
		panic(err)
	}

	defer interruptProcess(cmd)

	resp, err := http.Get("http://localhost:3000/internals")

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var appdef AppDef
	err = json.Unmarshal(body, &appdef)

	if err != nil {
		panic(err)
	}

	go serve(appdef, options)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	signalReceived := <-signalChanel

	switch signalReceived {
	case syscall.SIGINT, syscall.SIGTERM:
		log.Println("Received signal, shutting down...")
	default:
		log.Println("Received signal.")
		return 1
	}

	return 0
}

func serve(appdef AppDef, options options.HostOptions) {
	router := gin.Default()

	router.GET("/:funcName", func(c *gin.Context) {
		funcName := c.Param("funcName")

		for _, f := range appdef.Functions {
			if f == funcName {
				proxy(f, c.Writer, c.Request)
				return
			}
		}

		c.String(http.StatusNotFound, "Not found")
	})

	router.GET("/internals/options", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, options)
	})

	router.Run(":8080")
}

func proxy(endpoint string, writer http.ResponseWriter, req *http.Request) error {
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
