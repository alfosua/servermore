package main

import (
	"os"
	"os/signal"
)

func WaitSignals(sigs ...os.Signal) os.Signal {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, sigs...)
	return <-signalChanel
}
