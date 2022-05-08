package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"servermore/host/options"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ServermoreHost struct {
	Workers []*ServermoreWorker
	Options options.HostOptions
}

type ServermoreWorker struct {
	Id          string       `json:"guestId"`
	GuestEnv    string       `json:"guestEnv"`
	Application *Application `json:"app"`
}

func NewServermoreHost(options options.HostOptions) *ServermoreHost {
	return &ServermoreHost{
		Options: options,
	}
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
	worker := &ServermoreWorker{Id: workerId}
	err = json.Unmarshal(body, &worker)
	if err != nil {
		log.Fatal(err)
	}

	host.Workers = append(host.Workers, worker)
	log.Printf("Worker logged in: Id = %s, GuestEnv = %s\n", worker.Id, worker.GuestEnv)

	c.JSON(http.StatusCreated, *worker)
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
