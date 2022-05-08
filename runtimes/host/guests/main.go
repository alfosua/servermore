package guests

import (
	"fmt"
	"log"
	"servermore/host/options"
)

func StartWorker(ops options.HostOptions) (err error) {

	if ops.GuestEnv == "deno" {
		err = StartDenoWorker(ops.WorkerDir, ops.AppDir)
	} else {
		log.Fatal(
			fmt.Errorf("unknown guest environment: \"%s\"", ops.GuestEnv),
		)
	}

	return err
}
