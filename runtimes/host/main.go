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
	"github.com/google/uuid"
)

type ServermoreHost struct {
	Workers []*Worker
	Options options.HostOptions
}

type Worker struct {
	Id          string       `json:"guestId"`
	GuestEnv    string       `json:"guestEnv"`
	Application *Application `json:"app"`
}

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

func serve(host *ServermoreHost) {
	router := gin.Default()

	router.GET("/:funcName", host.ServerlessFunctionGet)

	router.GET("/internals/options", host.InternalOptionsGet)
	router.POST("/internals/worker", host.WorkerPost)

	router.Run(":8080")
}

func NewServermoreHost(options options.HostOptions) *ServermoreHost {
	return &ServermoreHost{
		Options: options,
	}
}

func (host *ServermoreHost) ServerlessFunctionGet(c *gin.Context) {
	funcName := c.Param("funcName")

	for _, w := range host.Workers {
		for _, f := range w.Application.Functions {
			if f == funcName {
				ReverseProxyOf(f, c.Writer, c.Request)
				return
			}
		}
	}

	c.String(http.StatusNotFound, "Not found")
}

func (host *ServermoreHost) InternalOptionsGet(c *gin.Context) {
	options := host.Options
	c.JSON(http.StatusOK, options)
}

func (host *ServermoreHost) WorkerPost(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
	}

	workerId := uuid.New().String()
	worker := &Worker{Id: workerId}
	err = json.Unmarshal(body, &worker)
	if err != nil {
		log.Fatal(err)
	}

	host.Workers = append(host.Workers, worker)
	log.Printf("Worker logged in: Id = %s, GuestEnv = %s\n", worker.Id, worker.GuestEnv)

	c.JSON(http.StatusCreated, *worker)
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
