package guests

import (
	"fmt"
	"os/exec"
	"servermore/host/options"
)

func StartWorker(ops options.HostOptions) (cmd *exec.Cmd, err error) {

	switch ops.GuestEnv {
	case "deno":
		cmd, err = StartDenoWorker(ops.WorkerDir, ops.AppDir)
	default:
		err = fmt.Errorf("unknown guest environment: %s", ops.GuestEnv)
	}

	return cmd, err
}
